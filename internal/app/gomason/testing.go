package gomason

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os"
	"os/exec"
)

// GovendorInstall  Installs govendor into the gopath indicated.
func GovendorInstall(gopath string, verbose bool) (err error) {
	if verbose {
		log.Printf("Installing Govendor with GOPATH=%s\n", gopath)
	}

	cmd := exec.Command("go", "get", "github.com/kardianos/govendor")

	runenv := append(os.Environ(), fmt.Sprintf("GOPATH=%s", gopath))

	cmd.Env = runenv

	err = cmd.Run()

	if err == nil {
		if verbose {
			log.Printf("Govendor installation complete\n\n")
		}
	}

	return err
}

// GovendorSync  Runs govendor sync in the diretory of your checked out code
func GovendorSync(gopath string, meta Metadata, verbose bool) (err error) {
	wd := fmt.Sprintf("%s/src/%s", gopath, meta.Package)

	if verbose {
		log.Printf("Changing working directory to: %s", wd)
	}

	err = os.Chdir(wd)

	if err != nil {
		log.Printf("Error changing working dir to %q: %s", wd, err)
		return err
	}

	govendor := fmt.Sprintf("%s/bin/govendor", gopath)

	var cmd *exec.Cmd

	if meta.InsecureGet {
		cmd = exec.Command(govendor, "sync", "-insecure")

	} else {
		cmd = exec.Command(govendor, "sync")

	}

	runenv := append(os.Environ(), fmt.Sprintf("GOPATH=%s", gopath))

	cmd.Env = runenv

	if verbose {
		cwd, err := os.Getwd()
		if err != nil {
			err = errors.Wrap(err, "failed to get working directory")
			return err
		}

		log.Printf("Working directory: %s", cwd)
		log.Printf("GOPATH: %s", gopath)
		log.Printf("Running %s sync", govendor)
	}

	err = cmd.Run()

	if err == nil {
		if verbose {
			log.Printf("Govendor sync complete.\n\n")
		}
	}

	return err
}

// GoTest  Runs 'go test -v ./...' in the checked out code directory
func GoTest(gopath string, gomodule string, verbose bool) (err error) {
	wd := fmt.Sprintf("%s/src/%s", gopath, gomodule)

	if verbose {
		log.Printf("Changing working directory to %s.\n", wd)
	}

	err = os.Chdir(wd)

	if err != nil {
		log.Printf("Error changing working dir to %q: %s", wd, err)
		return err
	}

	if verbose {
		log.Printf("Running 'go test -v ./...'.\n\n")
	}

	cmd := exec.Command("go", "test", "-v", "./...")

	runenv := append(os.Environ(), fmt.Sprintf("GOPATH=%s", gopath))

	cmd.Env = runenv

	output, err := cmd.CombinedOutput()

	log.Printf(string(output))

	if verbose {
		log.Printf("Done with go test.\n\n")
	}

	return err
}
