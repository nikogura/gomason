package gomason

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testMetaDataJson() string {
	return `{
	"version": "0.1.0",
	"package": "github.com/nikogura/testproject",
	"description": "Test Project for Gomason.",
	"building": {
		"prepcommands": [
      "echo \"GOPATH is: ${GOPATH}\""
		],
		"targets": [
			{
				"name": "linux/amd64"
			},
			{
				"name": "darwin/amd64"
			}
		]
	},
	"signing": {
		"program": "gpg",
		"email": "gomason-tester@foo.com"
	},
	"publishing": {
		"targets": [
			{
				"src": "testproject_darwin_amd64",
				"dst": "{{.Repository}}/testproject/{{.Version}}/darwin/amd64/testproject",
				"sig": true,
				"checksums": true
			},
			{
				"src": "testproject_linux_amd64",
				"dst": "{{.Repository}}/testproject/{{.Version}}/linux/amd64/testproject",
				"sig": true,
				"checksums": true
			}
		]
	}
}`
}

func testMetadataFileName() string {
	return "metadata.json"
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
	username, password, err := GetCredentials(testMetadataObj())
	if err != nil {
		log.Printf("Error getting credentials: %s", err)
		t.Fail()
	}

	matchUsername, err := regexp.MatchString(`.*`, username)
	if err != nil {
		log.Printf("Username fetch failed")
		t.Fail()
	}

	matchPassword, err := regexp.MatchString(`.*`, password)
	if err != nil {
		log.Printf("Password fetch failed")
		t.Fail()
	}

	assert.True(t, matchUsername, "Empty username")
	assert.True(t, matchPassword, "Empty password")
}

func TestGetFunc(t *testing.T) {
	expected := "foo"
	command := "echo 'foo'"

	actual, err := GetFunc(command)
	if err != nil {
		log.Printf("Error calling test func: %s", err)
		t.Fail()
	}

	assert.Equal(t, expected, actual, "Command output meets expectations.")

}
