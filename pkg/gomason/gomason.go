package gomason

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

// Metadata type to represent the metadata.json file
type Metadata struct {
	Version     string                 `json:"version"`
	Package     string                 `json:"package"`
	Description string                 `json:"description"`
	Repository  string                 `json:"repository"`
	InsecureGet bool                   `json:"insecure_get"`
	BuildInfo   BuildInfo              `json:"building,omitempty"`
	SignInfo    SignInfo               `json:"signing,omitempty"`
	PublishInfo PublishInfo            `json:"publishing,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

// BuildInfo stores information used for building the code.
type BuildInfo struct {
	Targets []string        `json:"targets,omitempty"`
	Extras  []ExtraArtifact `json:"extras,omitempty"`
}

// ExtraArtifact is an extra file built from a template at build time
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

// PublishInfo holds information for publishing
type PublishInfo struct {
	Targets      []PublishTarget          `json:"targets"`
	TargetsMap   map[string]PublishTarget `json:"-"`
	Username     string                   `json:"username"`
	Password     string                   `json:"password"`
	UsernameFunc string                   `json:"usernamefunc"`
	PasswordFunc string                   `json:"passwordfunc"`
	SkipSigning  bool                     `json:"skip-signing"`
}

// PublishTarget  a struct representing an individual file to upload
type PublishTarget struct {
	Source      string `json:"src"`
	Destination string `json:"dst"`
	Signature   bool   `json:"sig"`
	Checksums   bool   `json:"checksums"`
}

// UserConfig a struct representing the information stored in ~/.gomason
type UserConfig struct {
	User    UserInfo
	Signing UserSignInfo
}

// UserInfo  information from the user section in ~/.gomason
type UserInfo struct {
	Email        string
	Username     string
	Password     string
	UsernameFunc string
	PasswordFunc string
}

// UserSignInfo  information from the signing section in ~/.gomason
type UserSignInfo struct {
	Program string
}

// PublishBuildTargets loops over the expected files built by Build() and optionally signs them and publishes them along with their signatures (if signing).
//
// If not publishing, the binaries (and their optional signatures) are collected and dumped into the directory where gomason was called. (Typically the root of a go project).
func PublishBuildTargets(meta Metadata, gopath string, cwd string, sign bool, publish bool, collect bool, verbose bool) (err error) {
	// loop through the built things for each type of build target
	for _, arch := range meta.BuildInfo.Targets {
		if verbose {
			log.Printf("Processing build target: %s", arch)
		}
		archparts := strings.Split(arch, "/")

		osname := archparts[0]   // linux or darwin generally
		archname := archparts[1] // amd64 generally

		workdir := fmt.Sprintf("%s/src/%s", gopath, meta.Package)

		files, err := ioutil.ReadDir(workdir)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to read dir %s", workdir))
			return err
		}

		targetSuffix := fmt.Sprintf(".+_%s_%s", osname, archname)
		targetRegex := regexp.MustCompile(targetSuffix)

		for _, file := range files {
			matched := targetRegex.MatchString(file.Name())

			if matched {
				filename := fmt.Sprintf("%s/%s", workdir, file.Name())

				if _, err := os.Stat(filename); os.IsNotExist(err) {
					err = fmt.Errorf("failed to build binary: %s\n", filename)
					return err
				}

				// sign 'em if we're signing
				if sign {
					err = SignBinary(meta, filename, verbose)
					if err != nil {
						err = errors.Wrap(err, "failed to sign binary")
						return err
					}
				}

				// publish and return if we're publishing
				if publish {
					err = PublishFile(meta, filename, verbose)
					if err != nil {
						err = errors.Wrap(err, "failed to publish binary")
						return err
					}

				}

				if collect {
					// if we're not publishing, collect up the stuff we built, and dump 'em into the cwd where we called gomason
					err := CollectFileAndSignature(cwd, filename, verbose)
					if err != nil {
						err = errors.Wrap(err, "failed to collect binaries")
						return err
					}
				}
			}
		}

	}

	return err
}

// PublishBuildExtras loops over the expected files built by Build() and optionally signs them and publishes them along with their signatures (if signing).
//
// If not publishing, the binaries (and their optional signatures) are collected and dumped into the directory where gomason was called. (Typically the root of a go project).
func PublishBuildExtras(meta Metadata, gopath string, cwd string, sign bool, publish bool, verbose bool) (err error) {

	// loop through the built things for each type of build target
	for _, extra := range meta.BuildInfo.Extras {
		if verbose {
			log.Printf("Processing build extra: %s", extra.Template)
		}

		workdir := fmt.Sprintf("%s/src/%s", gopath, meta.Package)
		filename := fmt.Sprintf("%s/%s", workdir, extra.FileName)

		if _, err := os.Stat(filename); os.IsNotExist(err) {
			err = fmt.Errorf("failed to build extra artifact: %s\n", filename)
			return err
		}

		// sign 'em if we're signing
		if sign {
			err = SignBinary(meta, filename, verbose)
			if err != nil {
				err = errors.Wrap(err, "failed to sign extra artifact")
				return err
			}
		}

		// publish and return if we're publishing
		if publish {
			err = PublishFile(meta, filename, verbose)
			if err != nil {
				err = errors.Wrap(err, "failed to publish extra artifact")
				return err
			}

		} else {
			// if we're not publishing, collect up the stuff we built, and dump 'em into the cwd where we called gomason
			err := CollectFileAndSignature(cwd, filename, verbose)
			if err != nil {
				err = errors.Wrap(err, "failed to collect binaries")
				return err
			}
		}
	}

	return err
}

// CollectFileAndSignature grabs a file and the signature if it exists and moves it from the temp workspace into the CWD where gomason was called.
func CollectFileAndSignature(cwd string, filename string, verbose bool) (err error) {
	binaryDestinationPath := fmt.Sprintf("%s/%s", cwd, filepath.Base(filename))

	if verbose {
		log.Printf("Collecting Binaries and Signatures (if signing)")
	}

	err = os.Rename(filename, binaryDestinationPath)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to collect file %q", filename))
		return err
	}

	sigName := fmt.Sprintf("%s.asc", filepath.Base(filename))
	if _, err := os.Stat(sigName); !os.IsNotExist(err) {
		signatureDestinationPath := fmt.Sprintf("%s/%s", cwd, sigName)

		err = os.Rename(sigName, signatureDestinationPath)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to collect signature %q", sigName))
			return err
		}
	}

	return err
}

// GetUserConfig reads ~/.gomason if present, and returns a struct with its data.
func GetUserConfig() (config UserConfig, err error) {
	// pull per-user signing info out of ~/.gomason if present
	userObj, err := user.Current()
	if err != nil {
		err = fmt.Errorf("failed to get current user: %s", err)
		return config, err
	}

	homeDir := userObj.HomeDir

	if _, err := os.Stat(homeDir); os.IsNotExist(err) {
		err = fmt.Errorf("user %s's homedir %s does not exist", userObj.Name, homeDir)
		return config, err
	}

	perUserConfigFile := fmt.Sprintf("%s/.gomason", homeDir)

	if _, err := os.Stat(perUserConfigFile); !os.IsNotExist(err) {
		cfg, err := ini.Load(perUserConfigFile)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to load per user config file %s", perUserConfigFile))
			return config, err
		}

		userSection, _ := cfg.GetSection("user")
		if userSection != nil {
			userInfo := UserInfo{}

			// email section
			if userSection.HasKey("email") {
				key, _ := userSection.GetKey("email")
				userInfo.Email = key.Value()
			}

			// username section
			if userSection.HasKey("username") {
				key, _ := userSection.GetKey("username")
				userInfo.Username = key.Value()
			}

			// password section
			if userSection.HasKey("password") {
				key, _ := userSection.GetKey("password")
				userInfo.Password = key.Value()
			}

			// usernamefunc section
			if userSection.HasKey("usernamefunc") {
				key, _ := userSection.GetKey("usernamefunc")
				userInfo.UsernameFunc = key.Value()
			}

			// password func section
			if userSection.HasKey("passwordfunc") {
				key, _ := userSection.GetKey("passwordfunc")
				userInfo.PasswordFunc = key.Value()
			}

			config.User = userInfo
		}

		signingSection, _ := cfg.GetSection("signing")
		if signingSection != nil {
			signSec := UserSignInfo{}

			// program section
			if signingSection.HasKey("program") {
				key, _ := signingSection.GetKey("program")
				signSec.Program = key.Value()
			}

			config.Signing = signSec
		}
	}

	return config, err
}
