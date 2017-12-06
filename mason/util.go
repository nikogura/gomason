package mason

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

// CreateGoPath Creates an empty but workable GOPATH in the directory specified.  Returns the full GOPATH
func CreateGoPath(workDir string) (gopath string, err error) {
	gopath = fmt.Sprintf("%s/%s", workDir, "go")

	subdirs := []string{
		gopath,
		fmt.Sprintf("%s/%s", gopath, "src"),
		fmt.Sprintf("%s/%s", gopath, "bin"),
		fmt.Sprintf("%s/%s", gopath, "pkg"),
	}

	for _, dir := range subdirs {
		err = mkdir(dir, 0755)
		if err != nil {
			return gopath, err
		}
	}

	return gopath, err
}

func mkdir(dir string, perms os.FileMode) (err error) {
	err = os.MkdirAll(dir, perms)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to create dir %q: %s", dir, err))
		return err
	}

	return err
}

// ReadMetadata  Reads a metadata.json and returns the Metadata object thus described
func ReadMetadata(filename string) (metadata Metadata, err error) {
	mdBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return metadata, err
	}

	metadata.PublishInfo.Targets = make([]PublishTarget, 0)

	err = json.Unmarshal(mdBytes, &metadata)
	if err != nil {
		return metadata, err
	}

	// populate the targets map so we can look them up by src when publishing
	metadata.PublishInfo.TargetsMap = make(map[string]PublishTarget)
	for _, target := range metadata.PublishInfo.Targets {
		metadata.PublishInfo.TargetsMap[target.Source] = target
	}

	return metadata, err
}

// GitSSHUrlFromPackage Turns a go package name into a ssh git url.
func GitSSHUrlFromPackage(packageName string) (gitpath string) {
	munged := strings.Replace(packageName, "github.com/", "github.com:", 1)
	gitpath = fmt.Sprintf("git@%s.git", munged)

	return gitpath
}

// GetCredentials gets credentials, first from the metadata.json, and then from the user config in ~/.gomason if it exists.  If no credentials are found in any of the places, it returns the empty stings for usernames and passwords.  This is not recommended, but it might be useful in some cases.  Who knows?  We makes the tools, we don't tell you how to use them.  (we do, however make suggestions.) :D
func GetCredentials(meta Metadata, verbose bool) (username, password string, err error) {
	if verbose {
		log.Printf("Getting credentials")
	}

	// get creds from metadata
	// usernamefunc takes precedence over username
	if meta.PublishInfo.UsernameFunc != "" {
		log.Printf("Getting username from function")
		username, err = GetFunc(meta.PublishInfo.UsernameFunc, verbose)
		if err != nil {
			err = errors.Wrapf(err, "failed to get username from shell function %q", meta.PublishInfo.UsernameFunc)
			return username, password, err
		}
	} else if meta.PublishInfo.Username != "" {
		log.Printf("Getting username from metadata")
		username = meta.PublishInfo.Username
	}

	// passwordfunc takes precedence over password
	if meta.PublishInfo.PasswordFunc != "" {
		log.Printf("Getting password from function")
		password, err = GetFunc(meta.PublishInfo.PasswordFunc, verbose)
		if err != nil {
			err = errors.Wrapf(err, "failed to get password from shell function %q", meta.PublishInfo.PasswordFunc)
			return username, password, err
		}
	} else if meta.PublishInfo.Password != "" {
		log.Printf("Getting password from metadata")
		password = meta.PublishInfo.Password
	}

	// get creds from user config
	config, err := GetUserConfig()
	if err != nil {
		err = errors.Wrapf(err, "failed to get user config from ~/.gomason")
		return username, password, err
	}

	// usernamefunc takes precedence over username
	if config.User.UsernameFunc != "" {
		username, err = GetFunc(config.User.UsernameFunc, verbose)
		if err != nil {
			err = errors.Wrapf(err, "failed to get username from shell function %q", meta.PublishInfo.UsernameFunc)
			return username, password, err
		}
	} else if config.User.Username != "" {
		username = config.User.Username
	}

	// passwordfunc takes precedence over password
	if config.User.PasswordFunc != "" {
		password, err = GetFunc(config.User.PasswordFunc, verbose)
		if err != nil {
			err = errors.Wrapf(err, "failed to get password from shell function %q", meta.PublishInfo.UsernameFunc)
			return username, password, err
		}
	} else if config.User.Password != "" {
		password = config.User.Password
	}

	// We return empty strings for username and password if none is set anywhere.
	// The err variable will be nil in this case.  Why?  No creds configured is not necessarily an error.
	// People might want to use it without authentication
	// Not recommended, but we don't know where/how gomason will be used, and it might just make sense
	return username, password, err
}

// GetFunc runs a shell command that is a getter function.  This could certainly be dangerous, so be careful how you use it.
func GetFunc(shellCommand string, verbose bool) (result string, err error) {
	cmd := exec.Command("bash", "-c", shellCommand)

	if verbose {
		fmt.Printf("Getting input with shell function %q", shellCommand)
	}

	stdout, err := cmd.StdoutPipe()

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ()

	err = cmd.Start()
	if err != nil {
		err = errors.Wrapf(err, "failed to run %q", shellCommand)
		return result, err
	}

	stdoutBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		err = errors.Wrapf(err, "error reading stdout from func")
		return result, err
	}

	err = cmd.Wait()
	if err != nil {
		err = errors.Wrapf(err, "error waiting for %q to exit", shellCommand)
		return result, err
	}

	result = strings.TrimSuffix(string(stdoutBytes), "\n")

	return result, err
}

// ParseStringForMetadata parses a raw string as if it was a text/template template and uses the Metadata from metadata.json as it's data source.  e.g. injecting Version into upload targets (PUT url) when publishing.
func ParseStringForMetadata(rawUrlString string, metadata Metadata) (url string, err error) {
	tmpl, err := template.New("PublishingDestinationParse").Parse(rawUrlString)
	if err != nil {
		err = errors.Wrapf(err, "syntax error in destination url %q", rawUrlString)
		return url, err
	}

	buf := new(bytes.Buffer)

	err = tmpl.Execute(buf, metadata)
	if err != nil {
		err = errors.Wrapf(err, "failed to fill template %q with data", rawUrlString)
		return url, err
	}

	url = buf.String()

	return url, err
}
