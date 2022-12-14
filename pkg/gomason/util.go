package gomason

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

// AWS_ID_ENV_VAR Default env var for AWS access key
const AWS_ID_ENV_VAR = "AWS_ACCESS_KEY_ID"

// AWS_SECRET_ENV_VAR Default env var for AWS secret key
const AWS_SECRET_ENV_VAR = "AWS_SECRET_ACCESS_KEY"

// AWS_REGION_ENV_VAR Default env var for AWS region.
const AWS_REGION_ENV_VAR = "AWS_DEFAULT_REGION"

// ReadMetadata  Reads a metadata file and returns the Metadata object thus described
func ReadMetadata(filename string) (metadata Metadata, err error) {
	mdBytes, err := os.ReadFile(filename)
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

// GetCredentials gets credentials, first from the metadata file, and then from the user config in ~/.gomason if it exists.  If no credentials are found in any of the places, it returns the empty stings for usernames and passwords.  This is not recommended, but it might be useful in some cases.  Who knows?  We makes the tools, we don't tell you how to use them.  (we do, however make suggestions.) :D
func (g *Gomason) GetCredentials(meta Metadata) (username, password string, err error) {
	logrus.Debug("Getting credentials")

	// get creds from metadata
	// usernamefunc takes precedence over username
	if meta.PublishInfo.UsernameFunc != "" {
		logrus.Debug("Getting username from function")
		username, err = GetFunc(meta.PublishInfo.UsernameFunc)
		if err != nil {
			err = errors.Wrapf(err, "failed to get username from shell function %q", meta.PublishInfo.UsernameFunc)
			return username, password, err
		}
	} else if meta.PublishInfo.Username != "" {
		logrus.Debug("Getting username from metadata")
		username = meta.PublishInfo.Username
	}

	// passwordfunc takes precedence over password
	if meta.PublishInfo.PasswordFunc != "" {
		logrus.Debug("Getting password from function")
		password, err = GetFunc(meta.PublishInfo.PasswordFunc)
		if err != nil {
			err = errors.Wrapf(err, "failed to get password from shell function %q", meta.PublishInfo.PasswordFunc)
			return username, password, err
		}
	} else if meta.PublishInfo.Password != "" {
		logrus.Debug("Getting password from metadata")
		password = meta.PublishInfo.Password
	}

	config := g.Config

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

	stdoutBytes, err := io.ReadAll(stdout)
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

// ParseTemplateForMetadata parses a raw string as if it was a text/template template and uses the Metadata from metadata file as it's data source.  e.g. injecting Version into upload targets (PUT url) when publishing.
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

// DirsForURL given a URL, return a list of path elements suitable for creating directories/ folders
func DirsForURL(uri string) (dirs []string, err error) {
	dirs = make([]string, 0)

	u, err := url.Parse(uri)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse %s", uri)
		return dirs, err
	}

	dir := path.Dir(u.Path)
	dir = strings.TrimPrefix(dir, "/")
	parts := strings.Split(dir, "/")

	for len(parts) > 0 {
		dirs = append(dirs, strings.Join(parts, "/"))
		parts = parts[:len(parts)-1]
	}

	// Reverse the order, as this will be much easier to use the data to do path creation
	for i := len(dirs)/2 - 1; i >= 0; i-- {
		opp := len(dirs) - 1 - i
		dirs[i], dirs[opp] = dirs[opp], dirs[i]
	}

	return dirs, err
}

// DefaultSession creates a default AWS session from local config path.  Hooks directly into credentials if present, or Credentials Provider if configured.
func DefaultSession() (awssession *session.Session, err error) {
	if os.Getenv(AWS_ID_ENV_VAR) == "" && os.Getenv(AWS_SECRET_ENV_VAR) == "" {
		_ = os.Setenv("AWS_SDK_LOAD_CONFIG", "true")
	}

	awssession, err = session.NewSession()
	if err != nil {
		log.Fatalf("Failed to create aws session")
	}

	// For some reason this doesn't get picked up automatically, but we'll set it if it's present in the environment.
	if os.Getenv(AWS_REGION_ENV_VAR) != "" {
		awssession.Config.Region = aws.String(os.Getenv(AWS_REGION_ENV_VAR))
	}

	return awssession, err
}

// S3Meta a struct for holding metadata for S3 Objects.  There's probably already a struct that holds this, but this is all I need.
type S3Meta struct {
	Bucket string
	Region string
	Key    string
	Url    string
}

// S3Url returns true, and a metadata struct if the url given appears to be in s3
func S3Url(url string) (ok bool, meta S3Meta) {
	// Check to see if it's an s3 URL.
	s3Url := regexp.MustCompile(`https?://(.*)\.s3\.(.*)\.amazonaws.com/(.*)`)

	logrus.Debugf("testing %s", url)
	matches := s3Url.FindAllStringSubmatch(url, -1)

	if len(matches) == 0 {
		return ok, meta
	}

	match := matches[0]

	if len(match) == 4 {
		meta = S3Meta{
			Bucket: match[1],
			Region: match[2],
			Key:    match[3],
			Url:    url,
		}

		ok = true
		return ok, meta
	}

	return ok, meta
}
