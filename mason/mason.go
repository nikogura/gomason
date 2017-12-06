package mason

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strings"
)

// Metadata type to represent the metadata.json file
type Metadata struct {
	Version     string                 `json:"version"`
	Package     string                 `json:"package"`
	Description string                 `json:"description"`
	BuildInfo   BuildInfo              `json:"building,omitempty"`
	SignInfo    SignInfo               `json:"signing,omitempty"`
	PublishInfo PublishInfo            `json:"publishing,omitempty"`
	Options     map[string]interface{} `json:"-"`
}

// BuildInfo stores information used for building the code.
type BuildInfo struct {
	Targets []string `json:"targets,omitempty"`
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

// WholeShebang Creates an ephemeral workspace, installs Govendor into it, checks out your code, and runs the tests.  The whole shebang, hence the name.
//
// Optionally, it will build and publish your code too while it has the workspace set up.
//
// Specify workdir if you want to speed things up (govendor sync can take a while), but it's up to you to keep it clean.
//
// If workDir is the empty string, it will create and use a temporary directory.
func WholeShebang(workDir string, branch string, build bool, sign bool, publish bool, verbose bool) (err error) {
	var actualWorkDir string

	cwd, err := os.Getwd()
	if err != nil {
		err = errors.Wrap(err, "Failed to get current working directory.")
		return err
	}

	if workDir == "" {
		actualWorkDir, err = ioutil.TempDir("", "gomason")
		if err != nil {
			err = errors.Wrap(err, "Failed to create temp dir")
		}

		if verbose {
			log.Printf("Created temp dir %s", workDir)
		}

		defer os.RemoveAll(actualWorkDir)
	} else {
		actualWorkDir = workDir
	}

	gopath, err := CreateGoPath(actualWorkDir)
	if err != nil {
		return err
	}

	err = GovendorInstall(gopath, verbose)
	if err != nil {
		return err
	}

	meta, err := ReadMetadata("metadata.json")

	if err != nil {
		err = errors.Wrap(err, "couldn't read package information from metadata.json.")
		return err
	}

	err = Checkout(gopath, meta.Package, branch, verbose)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to checkout package %s at branch %s: %s", meta.Package, branch, err))
		return err
	}

	err = GovendorSync(gopath, meta.Package, verbose)
	if err != nil {
		err = errors.Wrap(err, "error running govendor sync")
		return err
	}

	err = GoTest(gopath, meta.Package, verbose)
	if err != nil {
		err = errors.Wrap(err, "error running go test")
		return err
	}

	log.Printf("Success!\n\n")

	if build {
		err = Build(gopath, meta.Package, branch, verbose)
		if err != nil {
			err = errors.Wrap(err, "build failed")
			return err
		}

		err = ProcessBuildTargets(meta, gopath, cwd, sign, publish, verbose)
		if err != nil {
			err = errors.Wrap(err, "post-build processing failed")
			return err
		}
	}
	return err
}

// ProcessBuildTargets loops over the expected files built by Build() and optionally signs them and publishes them along with their signatures (if signing).
//
// If not publishing, the binaries (and their optional signatures) are collected and dumped into the directory where gomason was called. (Typically the root of a go project).
func ProcessBuildTargets(meta Metadata, gopath string, cwd string, sign bool, publish bool, verbose bool) (err error) {
	parts := strings.Split(meta.Package, "/")
	binaryPrefix := parts[len(parts)-1]

	// loop through the built things for each type of build target
	for _, arch := range meta.BuildInfo.Targets {
		if verbose {
			log.Printf("Processing build target: %s", arch)
		}
		archparts := strings.Split(arch, "/")

		osname := archparts[0]   // linux or darwin generally
		archname := archparts[1] // amd64 generally

		workdir := fmt.Sprintf("%s/src/%s", gopath, meta.Package)
		binary := fmt.Sprintf("%s/%s_%s_%s", workdir, binaryPrefix, osname, archname)

		if _, err := os.Stat(binary); os.IsNotExist(err) {
			err = fmt.Errorf("Gox failed to build binary: %s\n", binary)
			return err
		}

		// sign 'em if we're signing
		if sign {
			err = SignBinary(meta, binary, verbose)
			if err != nil {
				err = errors.Wrap(err, "failed to sign binary")
				return err
			}
		}

		// publish and return if we're publishing
		if publish {
			err = PublishFile(meta, binary, verbose)
			if err != nil {
				err = errors.Wrap(err, "failed to publish binary")
				return err
			}

		} else {
			// if we're not publishing, collect up the stuff we built, and dump 'em into the cwd where we called gomason
			err := CollectBinaryAndSignature(cwd, binary, binaryPrefix, osname, archname, verbose)
			if err != nil {
				err = errors.Wrap(err, "failed to collect binaries")
				return err
			}
		}
	}

	return err
}

// CollectBinaryAndSignature grabs the binary and the signature if it exists and moves it from the temp workspace into the CWD where gomason was called.
func CollectBinaryAndSignature(cwd string, binary string, binaryPrefix string, osname string, archname string, verbose bool) (err error) {
	binaryDestinationPath := fmt.Sprintf("%s/%s_%s_%s", cwd, binaryPrefix, osname, archname)

	if verbose {
		log.Printf("Collecting Binaries and Signatures (if signing)")
	}

	err = os.Rename(binary, binaryDestinationPath)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to collect binary %q", binary))
		return err
	}

	sigName := fmt.Sprintf("%s.asc", binary)
	if _, err := os.Stat(sigName); !os.IsNotExist(err) {
		signatureDestinationPath := fmt.Sprintf("%s/%s_%s_%s.asc", cwd, binaryPrefix, osname, archname)

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
