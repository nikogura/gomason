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

	expected := TestMetadataObj()

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
	meta, err := ReadMetadata("../../metadata.json")
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
	username, password, err := GetCredentials(TestMetadataObj(), true)
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

	actual, err := GetFunc(command, true)
	if err != nil {
		log.Printf("Error calling test func: %s", err)
		t.Fail()
	}

	assert.Equal(t, expected, actual, "Command output meets expectations.")

}
