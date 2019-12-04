package languages

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/a8m/envsubst"
	"github.com/pkg/errors"

	"github.com/nikogura/gomason/pkg/gomason"
)

func init() {
	languagesMap["golang"] = Golang{}
}

// For golang, workdir is GOPATH
type Golang struct{}

// Creates an empty but workable GOPATH in the directory specified. Returns
// the full GOPATH
func (Golang) CreateWorkDir(workDir string) (gopath string, err error) {
	gopath = filepath.Join(workDir, "go")

	subdirs := []string{
		filepath.Join(gopath, "src"),
		filepath.Join(gopath, "bin"),
		filepath.Join(gopath, "pkg"),
	}

	for _, dir := range subdirs {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return gopath, err
		}
	}

	return gopath, err
}

// Checkout  Actually checks out the code you're trying to test into your temporary GOPATH
func (Golang) Checkout(gopath string, meta gomason.Metadata, branch string, verbose bool) (err error) {
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
		cmd = exec.Command(gocommand, "get", "-v", "-insecure", meta.Package)
	} else {
		cmd = exec.Command(gocommand, "get", "-v", "-d", fmt.Sprintf("%s/...", meta.Package))
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

	codepath := filepath.Join(gopath, "src", meta.Package)

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
func (Golang) Prep(gopath string, meta gomason.Metadata, verbose bool) (err error) {
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

// Runs 'go test -v ./...' in the checked out code directory
func (Golang) Test(gopath string, gomodule string, verbose bool) (err error) {
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

func (g Golang) Build(gopath string, meta gomason.Metadata, branch string, verbose bool) (err error) {
	if verbose {
		log.Printf("Checking to see that gox is installed.\n")
	}
	// Install gox if it's not already there
	if _, err := os.Stat(filepath.Join(gopath, "bin/gox")); os.IsNotExist(err) {
		err = GoxInstall(gopath, verbose)
		if err != nil {
			err = errors.Wrap(err, "Failed to install gox")
			return err
		}
	}

	if _, err := os.Stat(fmt.Sprintf("%s/src/%s/metadata.json", gopath, meta.Package)); os.IsNotExist(err) {
		err = g.Checkout(gopath, meta, branch, verbose)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("Failed to checkout module: %s branch: %s ", meta.Package, branch))
			return err
		}
	}

	wd := fmt.Sprintf("%s/src/%s", gopath, meta.Package)

	if verbose {
		log.Printf("Changing working directory to: %s", wd)
	}

	err = os.Chdir(wd)

	if err != nil {
		log.Printf("Error changing working dir to %q: %s", wd, err)
		return err
	}

	gox := fmt.Sprintf("%s/bin/gox", gopath)

	if verbose {
		log.Printf("Gox is: %s", gox)
	}

	metadatapath := fmt.Sprintf("%s/src/%s/metadata.json", gopath, meta.Package)

	md, err := gomason.ReadMetadata(metadatapath)
	if err != nil {
		err = errors.Wrap(err, "Failed to read metadata.json from checked out code")
		return err
	}

	for _, target := range md.BuildInfo.Targets {
		if verbose {
			log.Printf("Building target: %q\n", target.Name)
		}

		// This gets weird because go's exec shell doesn't like the arg format that gox expects
		// Building it thusly keeps the various quoting levels straight

		gopathenv := fmt.Sprintf("GOPATH=%s", gopath)
		runenv := append(os.Environ(), gopathenv)

		// allow user to turn off go modules
		if !target.Legacy {
			runenv = append(runenv, "GO111MODULE=on")
		}

		cgo := ""
		// build with cgo if we're told to do so.
		if target.Cgo {
			cgo = " -cgo"
		}

		for k, v := range target.Flags {
			runenv = append(runenv, fmt.Sprintf("%s=%s", k, v))
			if verbose {
				log.Printf("Build Flag: %s=%s", k, v)
			}
		}

		args := gox + cgo + ` -osarch="` + target.Name + `"` + " ./..."

		// Calling it through sh makes everything happy
		cmd := exec.Command("sh", "-c", args)

		cmd.Env = runenv

		if verbose {
			log.Printf("Running gox with: %s", args)
		}

		out, err := cmd.CombinedOutput()

		log.Printf("%s\n", string(out))

		if err != nil {
			log.Printf("Build error: %s\n", err.Error())
			return err
		}

		if verbose {
			log.Printf("Gox build of target %s complete and successful.\n\n", target.Name)
		}

	}

	err = BuildExtras(md, wd, verbose)
	if err != nil {
		err = errors.Wrapf(err, "Failed to build extras")
		return err

	}

	return err
}

// GoxInstall Installs github.com/mitchellh/gox, the go cross compiler
func GoxInstall(gopath string, verbose bool) (err error) {
	if verbose {
		log.Printf("Installing gox with GOPATH=%s\n", gopath)
	}

	gocommand, err := exec.LookPath("go")
	if err != nil {
		err = errors.Wrap(err, "Failed to find go binary")
		return err
	}

	cmd := exec.Command(gocommand, "get", "-v", "github.com/mitchellh/gox")

	env := append(os.Environ(), fmt.Sprintf("GOPATH=%s", gopath))

	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err == nil {
		if verbose {
			log.Printf("Gox successfully installed.\n\n")
		}
	}

	return err
}

// BuildExtras builds the extra artifacts specified in the metadata.json
func BuildExtras(meta gomason.Metadata, workdir string, verbose bool) (err error) {
	if verbose {
		log.Printf("Building Extra Artifacts")
	}

	for _, extra := range meta.BuildInfo.Extras {
		templateName := filepath.Join(workdir, extra.Template)
		outputFileName := filepath.Join(workdir, extra.FileName)
		executable := extra.Executable

		if verbose {
			fmt.Printf("Reading template from %s\n", templateName)
			fmt.Printf("Writing to %s\n", outputFileName)
		}

		var mode os.FileMode

		if executable {
			mode = 0755
		} else {
			mode = 0644
		}

		tmplBytes, err := ioutil.ReadFile(templateName)
		if err != nil {
			err = errors.Wrapf(err, "failed to read template file %s", templateName)
			return err
		}

		output, err := gomason.ParseTemplateForMetadata(string(tmplBytes), meta)
		if err != nil {
			err = errors.Wrapf(err, "failed to inject metadata into template text")
			return err
		}

		err = ioutil.WriteFile(outputFileName, []byte(output), mode)
		if err != nil {
			err = errors.Wrapf(err, "failed to write file %s", outputFileName)
			return err
		}
	}

	return err
}
