package gomason

import (
	"fmt"
	"github.com/a8m/envsubst"
	"github.com/pkg/errors"
	"log"
	"os"
	"os/exec"
)

// Checkout  Actually checks out the code you're trying to test into your temporary GOPATH
func Checkout(gopath string, meta Metadata, branch string, verbose bool) (err error) {

	err = os.Chdir(gopath)
	if err != nil {
		err = errors.Wrapf(err, "failed to cwd to %s", gopath)
		return err
	}

	// install the code via go get  after all, we don't really want to play if it's not in a repo.
	gocommand, err := exec.LookPath("go")
	if err != nil {
		err = errors.Wrap(err, "failed to find go binary")
		return err
	}

	runenv := append(os.Environ(), fmt.Sprintf("GOPATH=%s", gopath))
	runenv = append(runenv, "GO111MODULE=off")

	var cmd *exec.Cmd

	if meta.InsecureGet {
		cmd = exec.Command(gocommand, "get", "-insecure", meta.Package)

	} else {
		cmd = exec.Command(gocommand, "get", "-d", fmt.Sprintf("%s/...", meta.Package))
	}

	if verbose {
		log.Printf("Running %s with GOPATH=%s", cmd.Args, gopath)
	}

	cmd.Env = runenv

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	if err == nil {
		if verbose {
			log.Printf("Checkout of %s complete\n\n", meta.Package)
		}
	}

	git, err := exec.LookPath("git")
	if err != nil {
		err := errors.Wrap(err, "Failed to find git executable in path")
		return err
	}

	codepath := fmt.Sprintf("%s/src/%s", gopath, meta.Package)

	err = os.Chdir(codepath)

	if err != nil {
		log.Printf("Error changing working dir to %q: %s", codepath, err)
		return err
	}

	if branch != "" {
		if verbose {
			log.Printf("Checking out branch: %s\n\n", branch)
		}

		cmd := exec.Command(git, "checkout", branch)

		err = cmd.Run()

		if err == nil {
			if verbose {
				log.Printf("Checkout of branch: %s complete.\n\n", branch)
			}
		}
	}

	return err
}

// Prep  Commands run pre-build/ pre-test the checked out code in your temporary GOPATH
func Prep(gopath string, meta Metadata, verbose bool) (err error) {
	if verbose {
		log.Print("Running Prep Commands")
	}
	codepath := fmt.Sprintf("%s/src/%s", gopath, meta.Package)

	err = os.Chdir(codepath)
	if err != nil {
		err = errors.Wrapf(err, "failed to cwd to %s", gopath)
		return err
	}

	// set the gopath in the environment so that we can interpolate it below
	os.Setenv("GOPATH", gopath)

	for _, cmdString := range meta.BuildInfo.PrepCommands {

		// interpolate any environment variables into the command string
		cmdString, err = envsubst.String(cmdString)
		if err != nil {
			err = errors.Wrap(err, "failed to substitute env vars")
			return err
		}

		cmd := exec.Command("bash", "-c", cmdString)

		if verbose {
			log.Printf("Running %q with GOPATH=%s", cmdString, gopath)
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()

		if err != nil {
			err = errors.Wrapf(err, "failed running %q", cmdString)
			return err
		}
	}

	if err == nil {
		if verbose {
			log.Printf("Prep steps for %s complete\n\n", meta.Package)
		}
	}

	return err
}
