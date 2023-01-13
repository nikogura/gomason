package gomason

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/a8m/envsubst"
	"github.com/pkg/errors"
)

func init() {
	languagesMap[LanguageGolang] = Golang{}
}

// Golang struct.  For golang, workdir is GOPATH
type Golang struct{}

// CreateWorkDir Creates an empty but workable GOPATH in the directory specified. Returns
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
			err = errors.Wrapf(err, "failed creating directory")

			return gopath, err
		}
	}

	return gopath, err
}

// Checkout  Actually checks out the code you're trying to test into your temporary GOPATH
func (Golang) Checkout(gopath string, meta Metadata, branch string) (err error) {
	err = os.Chdir(gopath)
	if err != nil {
		err = errors.Wrapf(err, "failed to cwd to %s", gopath)
		return err
	}

	// install the code via go get  after all, we don't really want to play if it's not in a repo.
	gobinary := "go"
	gocommand, err := exec.LookPath(gobinary)
	if err != nil {
		err = errors.Wrapf(err, "failed to find go binary: %s", gobinary)
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

	logrus.Debugf("Running %s with GOPATH=%s", cmd.Args, gopath)

	cmd.Env = runenv

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	if err == nil {
		logrus.Debugf("Checkout of %s complete", meta.Package)
	}

	git, err := exec.LookPath("git")
	if err != nil {
		err := errors.Wrap(err, "Failed to find git executable in path")
		return err
	}

	codepath := filepath.Join(gopath, "src", meta.Package)

	err = os.Chdir(codepath)
	if err != nil {
		err = errors.Wrapf(err, "changing working dir to %q", codepath)
		return err
	}

	if branch != "" {
		logrus.Debugf("Checking out branch: %s", branch)

		cmd := exec.Command(git, "checkout", branch)

		err = cmd.Run()
		if err == nil {
			logrus.Debugf("Checkout of branch: %s complete.", branch)
		}
	}

	return err
}

// Prep  Commands run pre-build/ pre-test the checked out code in your temporary GOPATH
func (Golang) Prep(gopath string, meta Metadata) (err error) {
	logrus.Debug("Running Prep Commands")
	codepath := fmt.Sprintf("%s/src/%s", gopath, meta.Package)

	err = os.Chdir(codepath)
	if err != nil {
		err = errors.Wrapf(err, "failed to cwd to %s", gopath)
		return err
	}

	// set the gopath in the environment so that we can interpolate it below
	_ = os.Setenv("GOPATH", gopath)

	for _, cmdString := range meta.BuildInfo.PrepCommands {
		// interpolate any environment variables into the command string
		cmdString, err = envsubst.String(cmdString)
		if err != nil {
			err = errors.Wrap(err, "failed to substitute env vars")
			return err
		}

		cmd := exec.Command("bash", "-c", cmdString)

		logrus.Debugf("Running %q with GOPATH=%s", cmdString, gopath)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()

		if err != nil {
			err = errors.Wrapf(err, "failed running %q", cmdString)
			return err
		}
	}

	logrus.Debugf("Prep steps for %s complete", meta.Package)

	return err
}

// Test Runs 'go test -v ./...' in the checked out code directory
func (Golang) Test(gopath string, gomodule string, timeout string, local bool) (err error) {
	if !local {
		wd := filepath.Join(gopath, "src", gomodule)

		logrus.Debugf("Changing working directory to %s.", wd)

		err = os.Chdir(wd)

		if err != nil {
			err = errors.Wrapf(err, "changing working dir to %q", wd)
			return err
		}
	}

	logrus.Debugf("Running 'go test -v ./...'.")

	// TODO Should this use a shell exec like build?
	var cmd *exec.Cmd
	// Things break if you pass in an arg that has an empty string.  Splitting it up like this fixes https://github.com/nikogura/gomason/issues/24
	if timeout != "" {
		cmd = exec.Command("go", "test", "-v", "-timeout", timeout, "./...")
	} else {
		cmd = exec.Command("go", "test", "-v", "./...")
	}

	runenv := append(os.Environ(), fmt.Sprintf("GOPATH=%s", gopath))
	runenv = append(runenv, "GO111MODULE=on")

	cmd.Env = runenv

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Run()
	if err != nil {
		err = errors.Wrapf(err, "failed running %s", cmd)
	}

	logrus.Debugf("Done with go test.")

	return err
}

// Build uses `gox` to build binaries per metadata file
func (g Golang) Build(gopath string, meta Metadata, skipTargets string, local bool) (err error) {
	logrus.Debugf("Checking to see that gox is installed.")

	// Install gox if it's not already there
	if _, err := os.Stat(filepath.Join(gopath, "bin/gox")); os.IsNotExist(err) {
		err = GoxInstall(gopath)
		if err != nil {
			err = errors.Wrap(err, "Failed to install gox")
			return err
		}
	}

	var wd string

	if local {
		wd, err = os.Getwd()
		if err != nil {
			err = errors.Wrapf(err, "failed getting CWD")
			return err
		}
	} else {
		wd = fmt.Sprintf("%s/src/%s", gopath, meta.Package)

		logrus.Debugf("Changing working directory to: %s", wd)

		err = os.Chdir(wd)

		if err != nil {
			err = errors.Wrapf(err, "changing working dir to %q", wd)
			return err
		}
	}

	gox := fmt.Sprintf("%s/bin/gox", gopath)

	logrus.Debugf("Gox is: %s", gox)

	var metadatapath string
	if local {
		metadatapath = fmt.Sprintf("%s/%s", wd, METADATA_FILENAME)

	} else {
		metadatapath = fmt.Sprintf("%s/src/%s/%s", gopath, meta.Package, METADATA_FILENAME)
	}

	md, err := ReadMetadata(metadatapath)
	if err != nil {
		err = errors.Wrap(err, "Failed to read metadata file from checked out code")
		return err
	}

	skipTargetsMap := make(map[string]int)

	if skipTargets != "" {
		targetsList := strings.Split(skipTargets, ",")

		for _, t := range targetsList {
			skipTargetsMap[t] = 1
		}
	}

	for _, target := range md.BuildInfo.Targets {
		// skip this target if we're told to do so
		_, skip := skipTargetsMap[target.Name]
		if skip {
			continue
		}

		logrus.Debugf("Building target: %q in dir %s", target.Name, wd)

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
			logrus.Debugf("Build Flag: %s=%s", k, v)
		}

		ldflags := ""
		if target.Ldflags != "" {
			ldflags = fmt.Sprintf(" -ldflags %q ", target.Ldflags)
			logrus.Debugf("LD Flag: %s", ldflags)
		}

		args := gox + cgo + ldflags + ` -osarch="` + target.Name + `"` + " ./..."

		logrus.Debugf("Running gox with: %s in dir %s", args, wd)

		// Calling it through sh makes everything happy
		cmd := exec.Command("sh", "-c", args)

		cmd.Env = runenv

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		err = cmd.Run()
		if err != nil {
			err = errors.Wrapf(err, "failed building target %s", target.Name)
			return err
		}

		logrus.Debugf("Gox build of target %s complete and successful.", target.Name)
	}

	err = BuildExtras(md, wd)
	if err != nil {
		err = errors.Wrapf(err, "Failed to build extras")
		return err
	}

	return err
}

// GoxInstall Installs github.com/mitchellh/gox, the go cross compiler.
func GoxInstall(gopath string) (err error) {
	logrus.Debugf("Installing gox with GOPATH=%s, GOBIN=%s/bin", gopath, gopath)

	gocommand, err := exec.LookPath("go")
	if err != nil {
		err = errors.Wrap(err, "Failed to find go binary")
		return err
	}

	cmd := exec.Command(gocommand, "install", "-v", "github.com/mitchellh/gox@latest")

	env := append(os.Environ(), fmt.Sprintf("GOPATH=%s", gopath))
	env = append(os.Environ(), fmt.Sprintf("GOBIN=%s/bin", gopath))

	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	wd, err := os.Getwd()
	if err != nil {
		err = errors.Wrapf(err, "Error getting current working directory")
		return err
	}

	err = os.Chdir(gopath)
	if err != nil {
		err = errors.Wrapf(err, "Error changing directory into %s", gopath)
		return err
	}

	err = cmd.Run()
	if err != nil {
		err = errors.Wrapf(err, "failed installing gox")
		return err
	}

	goxPath := filepath.Join(gopath, "bin/gox")

	if _, err := os.Stat(goxPath); os.IsNotExist(err) {
		err = errors.New(fmt.Sprintf("Gox still not installed to %s", goxPath))
		return err
	}

	err = os.Chdir(wd)
	if err != nil {
		err = errors.Wrapf(err, "Error returning to directory %s", wd)
		return err
	}

	return err
}

// BuildExtras builds the extra artifacts specified in the metadata file
func BuildExtras(meta Metadata, workdir string) (err error) {
	logrus.Debugf("Building Extra Artifacts")

	for _, extra := range meta.BuildInfo.Extras {
		templateName := filepath.Join(workdir, extra.Template)
		outputFileName := filepath.Join(workdir, extra.FileName)
		executable := extra.Executable

		logrus.Debugf("Reading template from %s", templateName)
		logrus.Debugf("Writing to %s", outputFileName)

		var mode os.FileMode

		if executable {
			mode = 0755
		} else {
			mode = 0644
		}

		tmplBytes, err := os.ReadFile(templateName)
		if err != nil {
			err = errors.Wrapf(err, "failed to read template file %s", templateName)
			return err
		}

		output, err := ParseTemplateForMetadata(string(tmplBytes), meta)
		if err != nil {
			err = errors.Wrapf(err, "failed to inject metadata into template text")
			return err
		}

		err = os.WriteFile(outputFileName, []byte(output), mode)
		if err != nil {
			err = errors.Wrapf(err, "failed to write file %s", outputFileName)
			return err
		}
	}

	return err
}
