package mason

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"strings"
)

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

func GitSSHUrlFromPackage(packageName string) (gitpath string) {
	munged := strings.Replace(packageName, "github.com/", "github.com:", 1)
	gitpath = fmt.Sprintf("git@%s.git", munged)

	return gitpath
}
