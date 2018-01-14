package mason

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestCreateGoPath(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "gomason")
	if err != nil {
		log.Printf("Error creating temp dir\n")
		t.Fail()
	}

	defer os.RemoveAll(tmpDir)

	_, err = CreateGoPath(tmpDir)
	if err != nil {
		log.Printf("Error creating gopath in %q: %s", tmpDir, err)
		t.Fail()
	}

	dirs := []string{"go", "go/src", "go/pkg", "go/bin"}

	for _, dir := range dirs {
		if _, err := os.Stat(fmt.Sprintf("%s/%s", tmpDir, dir)); os.IsNotExist(err) {
			t.Fail()
		}
	}
}

func TestReadMetadata(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "gomason")
	if err != nil {
		log.Printf("Error creating temp dir\n")
		t.Fail()
	}

	defer os.RemoveAll(tmpDir)

	fileName := fmt.Sprintf("%s/%s", tmpDir, testMetadataFileName())

	err = ioutil.WriteFile(fileName, []byte(testMetaDataJson()), 0644)
	if err != nil {
		log.Printf("Error writing metadata file: %s", err)
		t.Fail()
	}

	expected := testMetadataObj()

	actual, err := ReadMetadata(fileName)
	if err != nil {
		log.Printf("Error reading metadata from file: %s", err)
		t.Fail()
	}

	assert.Equal(t, expected, actual, "Generated metadata object meets expectations.")

}

func TestGitSSHUrlFromPackage(t *testing.T) {
	input := "github.com/nikogura/gomason"
	expected := "git@github.com:nikogura/gomason.git"

	assert.Equal(t, expected, GitSSHUrlFromPackage(input), "Git SSH URL from Package Name meets expectations.")
}

func TestParseStringForMetadata(t *testing.T) {
	meta, err := ReadMetadata("metadata.json")
	if err != nil {
		log.Printf("Error reading metadata file: %s", err)
		t.Fail()
	}

	rawUrlString := testRawUrl()

	expected := testParsedUrl(meta.Version)

	actual, err := ParseTemplateForMetadata(rawUrlString, meta)
	if err != nil {
		log.Printf("Error parsing string: %s", err)
		t.Fail()
	}

	assert.Equal(t, expected, actual, "parsed url string meets expectations")
}

func TestGetCredentials(t *testing.T) {
	username, password, err := GetCredentials(testMetadataObj(), true)
	if err != nil {
		log.Printf("Error getting credentials: %s", err)
		t.Fail()
	}

	assert.Equal(t, "", username, "Empty username")
	assert.Equal(t, "", password, "Empty password")
}

func TestGetFunc(t *testing.T) {
	expected := "foo"
	command := "echo 'foo'"

	actual, err := GetFunc(command, true)
	if err != nil {
		log.Printf("Error calling test func: %s", err)
		t.Fail()
	}

	assert.Equal(t, expected, actual, "Command output meets expectations.")

}
