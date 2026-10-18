// Copyright © 2017 Nik Ogura <nik.ogura@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gomason

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

//nolint:gochecknoinits // logger configuration
func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

// Version returns the current gomason version from the embedded build info.
func Version() (version string) {
	version = "(unknown)"

	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Version != "" {
		version = info.Main.Version
	}

	return version
}

// MetadataFilename is the default gomason metadata file name.
const MetadataFilename = "metadata.json"

// Metadata is a type to represent the metadata file.
type Metadata struct {
	Name           string                 `json:"name"`
	Version        string                 `json:"version"`
	Package        string                 `json:"package"`
	Description    string                 `json:"description"`
	Repository     string                 `json:"repository"`
	ToolRepository string                 `json:"tool-repository"`
	InsecureGet    bool                   `json:"insecure_get"`
	Language       string                 `json:"language,omitempty"`
	BuildInfo      BuildInfo              `json:"building,omitempty"`
	SignInfo       SignInfo               `json:"signing,omitempty"`
	PublishInfo    PublishInfo            `json:"publishing,omitempty"`
	Options        map[string]interface{} `json:"options,omitempty"`
}

// GetLanguage returns the language set in metadata, or the default 'golang'.
func (m Metadata) GetLanguage() (lang string) {
	lang = m.Language

	if lang == "" {
		lang = LanguageGolang
	}

	return lang
}

// Gomason is the object that does all the building.
type Gomason struct {
	Config UserConfig
}

// NewGomason creates a new Gomason object for the current user.
func NewGomason() (g *Gomason, err error) {
	var userObj *user.User

	userObj, err = user.Current()
	if err != nil {
		err = errors.Wrapf(err, "failed to get current user")
		return g, err
	}

	var config UserConfig

	config, err = GetUserConfig(userObj.HomeDir)
	if err != nil {
		err = errors.Wrap(err, "error getting user config")
		return g, err
	}

	g = &Gomason{
		Config: config,
	}

	return g, err
}

// BuildInfo stores information used for building the code.
type BuildInfo struct {
	PrepCommands []string        `json:"prepcommands,omitempty"`
	Targets      []BuildTarget   `json:"targets,omitempty"`
	Extras       []ExtraArtifact `json:"extras,omitempty"`
	// Attestations gates generation of signed in-toto attestations (SLSA provenance
	// and a CycloneDX SBOM) alongside each built binary. Opt-in; absent means the
	// existing behavior, no attestations.
	Attestations bool `json:"attestations,omitempty"`
}

// BuildTarget contains information on each build target.
type BuildTarget struct {
	Name    string            `json:"name"`
	Cgo     bool              `json:"cgo,omitempty"`
	Flags   map[string]string `json:"flags,omitempty"`
	Ldflags string            `json:"ldflags,omitempty"`
	Legacy  bool              `json:"legacy,omitempty"`
}

// ExtraArtifact is an extra file built from a template at build time.
type ExtraArtifact struct {
	Template   string `json:"template"`
	FileName   string `json:"filename"`
	Executable bool   `json:"executable"`
}

// SignInfo holds information used for signing your binaries.
type SignInfo struct {
	Program string `json:"program"`
	Email   string `json:"email"`
}

// PublishInfo holds information for publishing.
type PublishInfo struct {
	Targets         []PublishTarget          `json:"targets"`
	TargetsMap      map[string]PublishTarget `json:"-"`
	Username        string                   `json:"username"`
	Password        string                   `json:"password"`
	UsernameFunc    string                   `json:"usernamefunc"`
	PasswordFunc    string                   `json:"passwordfunc"`
	BearerToken     string                   `json:"bearertoken"`
	BearerTokenFunc string                   `json:"bearertokenfunc"`
	SkipSigning     bool                     `json:"skip-signing"`
}

// PublishTarget is a struct representing an individual file to upload.
type PublishTarget struct {
	Source      string `json:"src"`
	Destination string `json:"dst"`
	Signature   bool   `json:"sig"`
	Checksums   bool   `json:"checksums"`
	// Attestations gates uploading the attestation files (provenance, SBOM, and their
	// signatures) alongside this artifact, when they were generated.
	Attestations bool `json:"attestations,omitempty"`
}

// UserConfig is a struct representing the information stored in ~/.gomason.
type UserConfig struct {
	User    UserInfo
	Signing UserSignInfo
}

// UserInfo  information from the user section in ~/.gomason.
type UserInfo struct {
	Email           string
	Username        string
	Password        string
	UsernameFunc    string
	PasswordFunc    string
	BearerToken     string
	BearerTokenFunc string
}

// UserSignInfo holds information from the signing section in ~/.gomason.
type UserSignInfo struct {
	Program string
}

// resolveWorkDir resolves the working directory based on whether we are running locally or not.
func resolveWorkDir(gopath string, meta Metadata, local bool) (workdir string, err error) {
	if local {
		workdir, err = os.Getwd()
		if err != nil {
			err = errors.Wrapf(err, "failed getting CWD")
		}

		return workdir, err
	}

	workdir = fmt.Sprintf("%s/src/%s", gopath, meta.Package)

	return workdir, err
}

// processArtifactFile handles signing, publishing, and collecting a single artifact file.
func (g *Gomason) processArtifactFile(meta Metadata, filename string, target BuildTarget, sign bool, publish bool, collect bool, cwd string) (err error) {
	_, statErr := os.Stat(filename)
	if os.IsNotExist(statErr) {
		err = errors.Wrapf(statErr, "failed building binary: %s\n", filename)
		return err
	}

	// sign 'em if we're signing
	if sign {
		logrus.Debugf("Signing %s", filename)
		err = g.SignBinary(meta, filename)
		if err != nil {
			err = errors.Wrapf(err, "failed to sign binary %s", filename)
			return err
		}

		// Attestations are generated and signed with the binary: an unsigned attestation
		// is a forgeable text file, so none are ever emitted without their signatures.
		// Only built binaries get attestations (target is set); templated extras do not.
		if meta.BuildInfo.Attestations && target.Name != "" {
			err = g.HandleAttestations(meta, filename, target)
			if err != nil {
				err = errors.Wrapf(err, "failed to generate attestations for %s", filename)
				return err
			}
		}
	}

	// publish and return if we're publishing
	if publish {
		logrus.Debugf("Publishing %s", filename)
		err = g.PublishFile(meta, filename)
		if err != nil {
			err = errors.Wrap(err, "failed to publish binary")
			return err
		}
	}

	// Collect up the stuff we built, and dump 'em into the cwd where we called gomason
	if collect {
		logrus.Debugf("Collecting %s", filename)
		err = CollectFileAndSignature(cwd, filename)
		if err != nil {
			err = errors.Wrap(err, "failed to collect binaries")
			return err
		}

		err = collectAttestationFiles(cwd, filename)
		if err != nil {
			err = errors.Wrap(err, "failed to collect attestations")
			return err
		}
	}

	return err
}

// HandleAttestations generates the provenance and SBOM attestation files for a built
// binary and signs each with the same identity that signed the binary.
func (g *Gomason) HandleAttestations(meta Metadata, filename string, target BuildTarget) (err error) {
	err = g.GenerateAttestationFiles(meta, filename, target)
	if err != nil {
		return err
	}

	for _, attFile := range []string{fmt.Sprintf("%s.att.json", filename), fmt.Sprintf("%s.sbom.json", filename)} {
		logrus.Debugf("Signing %s", attFile)

		err = g.SignBinary(meta, attFile)
		if err != nil {
			err = errors.Wrapf(err, "failed to sign attestation %s", attFile)
			return err
		}
	}

	return err
}

// collectAttestationFiles copies whatever attestation files exist beside an artifact
// into the cwd, alongside the artifact and its signature. Attestations are opt-in; when
// none were generated there is nothing to collect and this is a no-op.
func collectAttestationFiles(cwd string, filename string) (err error) {
	for _, suffix := range []string{".att.json", ".att.json.asc", ".sbom.json", ".sbom.json.asc"} {
		attPath := fmt.Sprintf("%s%s", filename, suffix)

		_, statErr := os.Stat(attPath)
		if os.IsNotExist(statErr) {
			continue
		}

		if statErr != nil {
			err = errors.Wrapf(statErr, "failed to stat %s", attPath)
			return err
		}

		destination := filepath.Join(cwd, filepath.Base(attPath))

		err = copyAndRemove(attPath, destination)
		if err != nil {
			return err
		}
	}

	return err
}

// HandleArtifacts loops over the expected files built by Build() and optionally signs them and publishes them along with their signatures (if signing).
//
// If not publishing, the binaries (and their optional signatures) are collected and dumped into the directory where gomason was called. (Typically the root of a go project).
func (g *Gomason) HandleArtifacts(meta Metadata, gopath string, cwd string, sign bool, publish bool, collect bool, skipTargets string, local bool) (err error) {
	logrus.Debug("Handling Artifacts\n")
	// loop through the built things for each type of build target
	skipTargetsMap := make(map[string]int)

	if skipTargets != "" {
		targetsList := strings.Split(skipTargets, ",")

		for _, t := range targetsList {
			skipTargetsMap[t] = 1
		}
	}

	for _, target := range meta.BuildInfo.Targets {
		// skip this target if we're told to do so
		_, skip := skipTargetsMap[target.Name]
		if skip {
			continue
		}

		err = g.handleTargetArtifacts(meta, gopath, cwd, target, sign, publish, collect, local)
		if err != nil {
			return err
		}
	}

	return err
}

// handleTargetArtifacts processes artifacts for a single build target.
func (g *Gomason) handleTargetArtifacts(meta Metadata, gopath string, cwd string, target BuildTarget, sign bool, publish bool, collect bool, local bool) (err error) {
	logrus.Debugf("Processing build target: %s\n", target.Name)
	archparts := strings.Split(target.Name, "/")

	osname := archparts[0]
	archname := archparts[1]

	var workdir string

	workdir, err = resolveWorkDir(gopath, meta, local)
	if err != nil {
		return err
	}

	var files []os.DirEntry

	files, err = os.ReadDir(workdir)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to read dir %s", workdir))
		return err
	}

	targetSuffix := fmt.Sprintf(".+_%s_%s$", osname, archname)
	targetRegex := regexp.MustCompile(targetSuffix)

	for _, file := range files {
		matched := targetRegex.MatchString(file.Name())
		if !matched {
			continue
		}

		filename := fmt.Sprintf("%s/%s", workdir, file.Name())
		logrus.Debugf("Handling %s", filename)

		err = g.processArtifactFile(meta, filename, target, sign, publish, collect, cwd)
		if err != nil {
			return err
		}
	}

	return err
}

// HandleExtras loops over the expected files built by Build() and optionally signs them and publishes them along with their signatures (if signing).
//
// If not publishing, the binaries (and their optional signatures) are collected and dumped into the directory where gomason was called. (Typically the root of a go project).
func (g *Gomason) HandleExtras(meta Metadata, gopath string, cwd string, sign bool, publish bool, collect bool, local bool) (err error) {

	// loop through the built things for each type of build target
	for _, extra := range meta.BuildInfo.Extras {
		logrus.Debugf("Processing build extra: %s", extra.Template)

		var workdir string

		workdir, err = resolveWorkDir(gopath, meta, local)
		if err != nil {
			return err
		}

		filename := fmt.Sprintf("%s/%s", workdir, extra.FileName)

		_, statErr := os.Stat(filename)
		if os.IsNotExist(statErr) {
			err = errors.Wrapf(statErr, "failed to build extra artifact: %s", filename)
			return err
		}

		// Extras are templated artifacts, not built binaries: they carry no attestations.
		err = g.processArtifactFile(meta, filename, BuildTarget{}, sign, publish, collect, cwd)
		if err != nil {
			return err
		}
	}

	return err
}

// DebugPrint prints debug information about a file.
func DebugPrint(filename, tag string) {
	ts := time.Now().UnixNano()
	cwd, _ := os.Getwd()
	logrus.Debugf("%s Current working directory: %s\n", tag, cwd)

	_, err := os.Stat(filename)
	if err != nil {
		logrus.Debugf("%s File %s missing at %v", tag, filename, ts)
	} else {
		logrus.Debugf("%s File %s exists at %v", tag, filename, ts)
	}
}

// copyAndRemove copies a file from src to dst preserving permissions, then removes the source.
func copyAndRemove(src string, dst string) (err error) {
	var fileInfo os.FileInfo

	fileInfo, err = os.Stat(src)
	if err != nil {
		err = errors.Wrapf(err, "failed statting file %s", src)
		return err
	}

	var contents []byte

	contents, err = os.ReadFile(src)
	if err != nil {
		err = errors.Wrapf(err, "failed reading file %q", src)
		return err
	}

	err = os.WriteFile(dst, contents, fileInfo.Mode())
	if err != nil {
		err = errors.Wrapf(err, "failed writing file %s", dst)
		return err
	}

	err = os.Remove(src)
	if err != nil {
		err = errors.Wrapf(err, "failed removing file %s", src)
	}

	return err
}

// CollectFileAndSignature grabs a file and the signature if it exists and copies it from the temp workspace into the CWD where gomason was called. Does nothing at all if the file is currently in cwd.
func CollectFileAndSignature(cwd string, filename string) (err error) {
	logrus.Debugf("Collecting Files and Signatures")

	binaryDestinationPath := fmt.Sprintf("%s/%s", cwd, filepath.Base(filename))

	if binaryDestinationPath != filename {
		err = copyAndRemove(filename, binaryDestinationPath)
		if err != nil {
			return err
		}
	}

	// The signature lives beside the file it signs — look there, not in the process's
	// current directory. Resolving relative to the process cwd silently dropped
	// signatures for any caller that had not chdir'd into the workdir.
	sigPath := fmt.Sprintf("%s.asc", filename)

	_, sigStatErr := os.Stat(sigPath)
	if os.IsNotExist(sigStatErr) {
		return err
	}

	signatureDestinationPath := fmt.Sprintf("%s/%s", cwd, filepath.Base(sigPath))
	if signatureDestinationPath != sigPath {
		err = copyAndRemove(sigPath, signatureDestinationPath)
	}

	return err
}

// parseUserSection parses the user section from a gomason config file.
func parseUserSection(cfg *ini.File) (userInfo UserInfo) {
	userSection, _ := cfg.GetSection("user")
	if userSection == nil {
		return userInfo
	}

	if userSection.HasKey("email") {
		key, _ := userSection.GetKey("email")
		userInfo.Email = key.Value()
	}

	if userSection.HasKey("username") {
		key, _ := userSection.GetKey("username")
		userInfo.Username = key.Value()
	}

	if userSection.HasKey("password") {
		key, _ := userSection.GetKey("password")
		userInfo.Password = key.Value()
	}

	if userSection.HasKey("usernamefunc") {
		key, _ := userSection.GetKey("usernamefunc")
		userInfo.UsernameFunc = key.Value()
	}

	if userSection.HasKey("passwordfunc") {
		key, _ := userSection.GetKey("passwordfunc")
		userInfo.PasswordFunc = key.Value()
	}

	if userSection.HasKey("bearertoken") {
		key, _ := userSection.GetKey("bearertoken")
		userInfo.BearerToken = key.Value()
	}

	if userSection.HasKey("bearertokenfunc") {
		key, _ := userSection.GetKey("bearertokenfunc")
		userInfo.BearerTokenFunc = key.Value()
	}

	return userInfo
}

// parseSigningSection parses the signing section from a gomason config file.
func parseSigningSection(cfg *ini.File) (signInfo UserSignInfo) {
	signingSection, _ := cfg.GetSection("signing")
	if signingSection == nil {
		return signInfo
	}

	if signingSection.HasKey("program") {
		key, _ := signingSection.GetKey("program")
		signInfo.Program = key.Value()
	}

	return signInfo
}

// GetUserConfig reads ~/.gomason if present, and returns a struct with its data.
func GetUserConfig(homedir string) (config UserConfig, err error) {
	perUserConfigFile := fmt.Sprintf("%s/.gomason", homedir)

	_, statErr := os.Stat(perUserConfigFile)
	if os.IsNotExist(statErr) {
		return config, err
	}

	var cfg *ini.File

	cfg, err = ini.Load(perUserConfigFile)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to load per user config file %s", perUserConfigFile))
		return config, err
	}

	config.User = parseUserSection(cfg)
	config.Signing = parseSigningSection(cfg)

	return config, err
}
