package mason

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"strings"
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

	err = json.Unmarshal(mdBytes, &metadata)
	if err != nil {
		return metadata, err
	}

	return metadata, err
}

// GitSSHUrlFromPackage Turns a go package name into a ssh git url.
func GitSSHUrlFromPackage(packageName string) (gitpath string) {
	munged := strings.Replace(packageName, "github.com/", "github.com:", 1)
	gitpath = fmt.Sprintf("git@%s.git", munged)

	return gitpath
}
