package gomason

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/a8m/envsubst"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

//nolint:gochecknoinits // language registration pattern
func init() {
	languagesMap[LanguageGolang] = Golang{}
}

// Golang struct.  For golang, workdir is GOPATH.
type Golang struct{}

// CreateWorkDir Creates an empty but workable GOPATH in the directory specified. Returns
// the full GOPATH.
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

// Checkout actually checks out the code you're trying to test into your temporary GOPATH.
func (Golang) Checkout(gopath string, meta Metadata, branch string) (err error) {
	err = os.Chdir(gopath)
	if err != nil {
		err = errors.Wrapf(err, "failed to cwd to %s", gopath)
		return err
	}

	// Try git clone first as fallback for modern Go compatibility
	gitErr := checkoutViaGit(gopath, meta, branch)
	if gitErr == nil {
		logrus.Debugf("Git checkout of %s complete", meta.Package)
		return err
	}

	logrus.Debugf("Git checkout failed, trying go get: %v", gitErr)

	// Fallback to go get for legacy compatibility
	gobinary := "go"

	var gocommand string

	gocommand, err = exec.LookPath(gobinary)
	if err != nil {
		err = errors.Wrapf(err, "failed to find go binary: %s", gobinary)
		return err
	}

	runenv := append(os.Environ(), fmt.Sprintf("GOPATH=%s", gopath))
	// Try GO111MODULE=off for legacy compatibility
	runenv = append(runenv, "GO111MODULE=off")

	var cmd *exec.Cmd

	if meta.InsecureGet {
		cmd = exec.CommandContext(context.Background(), gocommand, "get", "-v", "-insecure", meta.Package)
	} else {
		cmd = exec.CommandContext(context.Background(), gocommand, "get", "-v", "-d", fmt.Sprintf("%s/...", meta.Package))
	}

	logrus.Debugf("Running %s with GOPATH=%s", cmd.Args, gopath)

	cmd.Env = runenv

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	if err == nil {
		logrus.Debugf("Checkout of %s complete", meta.Package)
	}

	var git string

	git, err = exec.LookPath("git")
	if err != nil {
		err = errors.Wrap(err, "Failed to find git executable in path")
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

		branchCmd := exec.CommandContext(context.Background(), git, "checkout", branch)

		err = branchCmd.Run()
		if err == nil {
			logrus.Debugf("Checkout of branch: %s complete.", branch)
		}
	}

	return err
}

// checkoutViaGit clones a repository using git directly, useful for modern Go projects.
func checkoutViaGit(gopath string, meta Metadata, branch string) (err error) {
	var git string

	git, err = exec.LookPath("git")
	if err != nil {
		err = errors.Wrap(err, "failed to find git executable in path")
		return err
	}

	// Create the target directory structure
	srcDir := filepath.Join(gopath, "src")
	err = os.MkdirAll(srcDir, 0755)
	if err != nil {
		err = errors.Wrapf(err, "failed to create src directory: %s", srcDir)
		return err
	}

	err = os.Chdir(srcDir)
	if err != nil {
		err = errors.Wrapf(err, "failed to change to src directory: %s", srcDir)
		return err
	}

	// Convert package path to git URL
	gitURL := fmt.Sprintf("https://%s.git", meta.Package)

	// Check if target directory already exists and remove it
	targetDir := filepath.Join(srcDir, meta.Package)

	_, statErr := os.Stat(targetDir)
	if statErr == nil {
		err = os.RemoveAll(targetDir)
		if err != nil {
			err = errors.Wrapf(err, "failed to remove existing directory: %s", targetDir)
			return err
		}
	}

	// Clone the repository
	var cmd *exec.Cmd
	if branch != "" {
		cmd = exec.CommandContext(context.Background(), git, "clone", "-b", branch, gitURL, meta.Package)
	} else {
		cmd = exec.CommandContext(context.Background(), git, "clone", gitURL, meta.Package)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		err = errors.Wrapf(err, "failed to clone repository: %s", gitURL)
		return err
	}

	return err
}

// Prep runs commands pre-build/pre-test on the checked out code in your temporary GOPATH.
func (Golang) Prep(gopath string, meta Metadata, local bool) (err error) {
	logrus.Debug("Running Prep Commands")
	if local {
		_, err = os.Getwd()
		if err != nil {
			err = errors.Wrapf(err, "failed getting CWD")
			return err
		}

	} else {
		codepath := fmt.Sprintf("%s/src/%s", gopath, meta.Package)

		err = os.Chdir(codepath)
		if err != nil {
			err = errors.Wrapf(err, "failed to cwd to %s", gopath)
			return err
		}

		// set the gopath in the environment so that we can interpolate it below
		_ = os.Setenv("GOPATH", gopath)
	}

	for _, cmdString := range meta.BuildInfo.PrepCommands {
		// interpolate any environment variables into the command string
		cmdString, err = envsubst.String(cmdString)
		if err != nil {
			err = errors.Wrap(err, "failed to substitute env vars")
			return err
		}

		cmd := exec.CommandContext(context.Background(), "bash", "-c", cmdString)

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

// Test runs 'go test -v ./...' in the checked out code directory.
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
		cmd = exec.CommandContext(context.Background(), "go", "test", "-v", "-timeout", timeout, "./...")
	} else {
		cmd = exec.CommandContext(context.Background(), "go", "test", "-v", "./...")
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

// Build uses `gox` to build binaries per metadata file.
func (g Golang) Build(gopath string, meta Metadata, skipTargets string, local bool) (err error) {
	logrus.Debugf("Checking to see that gox is installed.")

	// Install gox if it's not already there
	_, statErr := os.Stat(filepath.Join(gopath, "bin/gox"))
	if os.IsNotExist(statErr) {
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
		metadatapath = fmt.Sprintf("%s/%s", wd, MetadataFilename)

	} else {
		metadatapath = fmt.Sprintf("%s/src/%s/%s", gopath, meta.Package, MetadataFilename)
	}

	var md Metadata

	md, err = ReadMetadata(metadatapath)
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

		err = buildSingleTarget(gox, gopath, target, wd, local)
		if err != nil {
			return err
		}
	}

	err = BuildExtras(md, wd)
	if err != nil {
		err = errors.Wrapf(err, "Failed to build extras")
		return err
	}

	return err
}

// buildSingleTarget builds a single gox target.
func buildSingleTarget(gox string, gopath string, target BuildTarget, wd string, local bool) (err error) {
	logrus.Debugf("Building target: %q in dir %s", target.Name, wd)

	// This gets weird because go's exec shell doesn't like the arg format that gox expects
	// Building it thusly keeps the various quoting levels straight

	runenv := os.Environ()

	if !local {
		gopathenv := fmt.Sprintf("GOPATH=%s", gopath)
		runenv = append(runenv, gopathenv)
	}

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

	// Interesting idea, but breaks multiple binary builds such as dbt.  To properly implement, we'd have to find and handle each binary instead of relying on the './...'.
	//outputTemplate := fmt.Sprintf("%s_{{.OS}}_{{.Arch}}", meta.Name)
	//args := gox + cgo + ldflags + ` -osarch="` + target.Name + `"` + ` -output="` + outputTemplate + `"` + " ./..."

	args := gox + cgo + ldflags + ` -osarch="` + target.Name + `"` + " ./..."

	logrus.Debugf("Running gox with: %s in dir %s", args, wd)

	// Calling it through sh makes everything happy
	cmd := exec.CommandContext(context.Background(), "sh", "-c", args)

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

	return err
}

// GoxInstall Installs github.com/mitchellh/gox, the go cross compiler.
func GoxInstall(gopath string) (err error) {
	logrus.Debugf("Installing gox with GOPATH=%s, GOBIN=%s/bin", gopath, gopath)

	var gocommand string

	gocommand, err = exec.LookPath("go")
	if err != nil {
		err = errors.Wrap(err, "Failed to find go binary")
		return err
	}

	cmd := exec.CommandContext(context.Background(), gocommand, "install", "-v", "github.com/mitchellh/gox@latest")

	env := append(os.Environ(), fmt.Sprintf("GOPATH=%s", gopath))
	env = append(env, fmt.Sprintf("GOBIN=%s/bin", gopath))

	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	var wd string

	wd, err = os.Getwd()
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

	_, statErr := os.Stat(goxPath)
	if os.IsNotExist(statErr) {
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

// BuildExtras builds the extra artifacts specified in the metadata file.
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

		var tmplBytes []byte

		tmplBytes, err = os.ReadFile(templateName)
		if err != nil {
			err = errors.Wrapf(err, "failed to read template file %s", templateName)
			return err
		}

		var output string

		output, err = ParseTemplateForMetadata(string(tmplBytes), meta)
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
