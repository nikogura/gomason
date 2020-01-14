package gomason

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

// ReadMetadata  Reads a metadata.json and returns the Metadata object thus described
func ReadMetadata(filename string) (metadata Metadata, err error) {
	mdBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		err = errors.Wrapf(err, "failed to read file %s", filename)
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
func GetCredentials(meta Metadata) (username, password string, err error) {
	log.Print("[DEBUG] Getting credentials")

	// get creds from metadata
	// usernamefunc takes precedence over username
	if meta.PublishInfo.UsernameFunc != "" {
		log.Printf("Getting username from function")
		username, err = GetFunc(meta.PublishInfo.UsernameFunc)
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
		password, err = GetFunc(meta.PublishInfo.PasswordFunc)
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
		username, err = GetFunc(config.User.UsernameFunc)
		if err != nil {
			err = errors.Wrapf(err, "failed to get username from shell function %q", meta.PublishInfo.UsernameFunc)
			return username, password, err
		}
	} else if config.User.Username != "" {
		username = config.User.Username
	}

	// passwordfunc takes precedence over password
	if config.User.PasswordFunc != "" {
		password, err = GetFunc(config.User.PasswordFunc)
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
func GetFunc(shellCommand string) (result string, err error) {
	cmd := exec.Command("sh", "-c", shellCommand)

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

// ParseTemplateForMetadata parses a raw string as if it was a text/template template and uses the Metadata from metadata.json as it's data source.  e.g. injecting Version into upload targets (PUT url) when publishing.
func ParseTemplateForMetadata(templateText string, metadata Metadata) (outputText string, err error) {
	tmpl, err := template.New("OnTheFlyTemplate").Parse(templateText)
	if err != nil {
		err = errors.Wrapf(err, "syntax error in destination url %q", templateText)
		return outputText, err
	}

	buf := new(bytes.Buffer)

	err = tmpl.Execute(buf, metadata)
	if err != nil {
		err = errors.Wrapf(err, "failed to fill template %q with data", templateText)
		return outputText, err
	}

	outputText = buf.String()

	return outputText, err
}
