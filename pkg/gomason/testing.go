package gomason

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

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
	runenv = append(runenv, "GO111MODULE=on")

	cmd.Env = runenv

	output, err := cmd.CombinedOutput()

	log.Printf(string(output))

	if verbose {
		log.Printf("Done with go test.\n\n")
	}

	return err
}
