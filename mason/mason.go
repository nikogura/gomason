package mason

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// Metadata type to represent the metadata.json file
type Metadata struct {
	Version      string                 `json:"version"`
	Package      string                 `json:"package"`
	Description  string                 `json:"description"`
	BuildTargets []string               `json:"buildtargets,omitempty"`
	Signing      Signing                `json:"signing,omitempty"`
	Options      map[string]interface{} `json:"-"`
}

//Signing information
type Signing struct {
	Program string `json:"program"`
	Email   string `json:"email"`
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

// ProcessBuildTargets loops over the expected binaries built by Build() and optionally signs them and publishes them along with their signatures (if signing).
//
// If not publishing, the binaries (and their optional signatures) are collected and dumped into the directory where gomason was called. (Typically the root of a go project).
func ProcessBuildTargets(meta Metadata, gopath string, cwd string, sign bool, publish bool, verbose bool) (err error) {
	parts := strings.Split(meta.Package, "/")
	binaryPrefix := parts[len(parts)-1]

	// loop through the built things for each type of build target
	for _, arch := range meta.BuildTargets {
		archparts := strings.Split(arch, "/")

		osname := archparts[0]   // linux or darwin generally
		archname := archparts[1] // amd64 generally

		workdir := fmt.Sprintf("%s/src/%s", gopath, meta.Package)
		binary := fmt.Sprintf("%s/%s_%s_%s", workdir, binaryPrefix, osname, archname)

		if _, err := os.Stat(binary); os.IsNotExist(err) {
			err = errors.New(fmt.Sprintf("Gox failed to build binary: %s\n", binary))
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
			err = PublishBinary(meta, cwd, binary, binaryPrefix, osname, archname)
			if err != nil {
				err = errors.Wrap(err, "failed to publish binary")
				return err
			}

			return err
		}

		// if we're not publishing, collect up the stuff we built, and dump 'em into the cwd where we called gomason
		err := CollectBinaryAndSignature(cwd, binary, binaryPrefix, osname, archname, verbose)
		if err != nil {
			err = errors.Wrap(err, "failed to collect binaries")
			return err
		}
	}

	return err
}

// CollectBinaryAndSignature grabs the binary and the signature if it exists and moves it from the temp workspace into the CWD where gomason was called.
func CollectBinaryAndSignature(cwd string, binary string, binaryPrefix string, osname string, archname string, verbose bool) (err error) {
	binaryDestinationPath := fmt.Sprintf("%s/%s_%s_%s", cwd, binaryPrefix, osname, archname)

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
