package mason

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os"
	"os/exec"
)

// Checkout  Actually checks out the code you're trying to test into your temporary GOPATH
func Checkout(gopath string, meta Metadata, branch string, verbose bool) (err error) {

	// install the code via go get  after all, we don't really want to play if it's not in a repo.
	gocommand, err := exec.LookPath("go")
	if err != nil {
		err = errors.Wrap(err, "failed to find go binary")
		return err
	}

	runenv := append(os.Environ(), fmt.Sprintf("GOPATH=%s", gopath))

	var cmd *exec.Cmd

	if meta.InsecureGet {
		cmd = exec.Command(gocommand, "get", "-insecure", meta.Package)

	} else {
		cmd = exec.Command(gocommand, "get", meta.Package)

	}

	if verbose {
		log.Printf("Running %s %s", gocommand, cmd.Args)
	}

	cmd.Env = runenv

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
