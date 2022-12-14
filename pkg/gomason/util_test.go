package gomason

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

func TestReadMetadata(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gomason")
	if err != nil {
		log.Printf("Error creating temp dir\n")
		t.Fail()
	}
	defer os.RemoveAll(tmpDir)

	fileName := fmt.Sprintf("%s/%s", tmpDir, testMetadataFileName())

	err = os.WriteFile(fileName, []byte(testMetaDataJson()), 0644)
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

//func TestGitSSHUrlFromPackage(t *testing.T) {
//	input := "github.com/nikogura/gomason"
//	expected := "git@github.com:nikogura/gomason.git"
//
//	assert.Equal(t, expected, GitSSHUrlFromPackage(input), "Git SSH URL from Package Name meets expectations.")
//}
//
//func TestParseStringForMetadata(t *testing.T) {
//	meta, err := ReadMetadata(METADATA_FILENAME)
//	if err != nil {
//		log.Printf("Error reading metadata file: %s", err)
//		t.Fail()
//	}
//
//	rawUrlString := testRawUrl()
//
//	expected := testParsedUrl(meta.Version)
//
//	actual, err := ParseTemplateForMetadata(rawUrlString, meta)
//	if err != nil {
//		log.Printf("Error parsing string: %s", err)
//		t.Fail()
//	}
//
//	assert.Equal(t, expected, actual, "parsed url string meets expectations")
//}
//
//func TestGetCredentials(t *testing.T) {
//	g := Gomason{
//		Config: UserConfig{
//			User:    UserInfo{},
//			Signing: UserSignInfo{},
//		},
//	}
//	username, password, err := g.GetCredentials(testMetadataObj())
//	if err != nil {
//		log.Printf("Error getting credentials: %s", err)
//		t.Fail()
//	}
//
//	matchUsername, err := regexp.MatchString(`.*`, username)
//	if err != nil {
//		log.Printf("Username fetch failed")
//		t.Fail()
//	}
//
//	matchPassword, err := regexp.MatchString(`.*`, password)
//	if err != nil {
//		log.Printf("Password fetch failed")
//		t.Fail()
//	}
//
//	assert.True(t, matchUsername, "Empty username")
//	assert.True(t, matchPassword, "Empty password")
//}
//
//func TestGetFunc(t *testing.T) {
//	expected := "foo"
//	command := "echo 'foo'"
//
//	actual, err := GetFunc(command)
//	if err != nil {
//		log.Printf("Error calling test func: %s", err)
//		t.Fail()
//	}
//
//	assert.Equal(t, expected, actual, "Command output meets expectations.")
//
//}
//
//func TestDefaultSession(t *testing.T) {
//	_, err := DefaultSession()
//	if err != nil {
//		t.Errorf("Failed to get an AWS Session")
//	}
//}
//
//func TestS3Url(t *testing.T) {
//	inputs := []struct {
//		url    string
//		result bool
//		bucket string
//		region string
//		key    string
//	}{
//		{
//			"https://www.nikogura.com",
//			false,
//			"",
//			"",
//			"",
//		},
//		{
//			"https://dbt-tools.s3.us-east-1.amazonaws.com/catalog/1.2.3/linux/amd64/catalog",
//			true,
//			"dbt-tools",
//			"us-east-1",
//			"catalog/1.2.3/linux/amd64/catalog",
//		},
//	}
//
//	for _, tc := range inputs {
//		t.Run(tc.url, func(t *testing.T) {
//			fmt.Printf("Testing %s\n", tc.url)
//			ok, meta := S3Url(tc.url)
//
//			assert.True(t, ok == tc.result, fmt.Sprintf("%s does not meet expectations", tc.url))
//			assert.True(t, tc.bucket == meta.Bucket, fmt.Sprintf("Bucket %q doesn't look right", meta.Bucket))
//			assert.True(t, tc.region == meta.Region, fmt.Sprintf("Region %q doesn't look right.", meta.Region))
//			assert.True(t, tc.key == meta.Key, fmt.Sprintf("Key %q doesn't look right.", meta.Key))
//		})
//	}
//}
//
//func TestDirsForURL(t *testing.T) {
//	inputs := []struct {
//		name   string
//		input  string
//		output []string
//	}{
//		{
//			"s3 reposerver url",
//			"https://foo.com/dbt-tools/catalog/1.2.3/linux/amd64/catalog",
//			[]string{
//				"dbt-tools",
//				"dbt-tools/catalog",
//				"dbt-tools/catalog/1.2.3",
//				"dbt-tools/catalog/1.2.3/linux",
//				"dbt-tools/catalog/1.2.3/linux/amd64",
//			},
//		},
//		{
//			"s3 catalog url",
//			"https://dbt-tools.s3.us-east-1.amazonaws.com/catalog/1.2.3/linux/amd64/catalog",
//			[]string{
//				"catalog",
//				"catalog/1.2.3",
//				"catalog/1.2.3/linux",
//				"catalog/1.2.3/linux/amd64",
//			},
//		},
//	}
//
//	for _, tc := range inputs {
//		t.Run(tc.name, func(t *testing.T) {
//			dirs, err := DirsForURL(tc.input)
//			if err != nil {
//				t.Error(err)
//			}
//
//			assert.Equal(t, tc.output, dirs, "Parsed directories meet expectations")
//		})
//	}
//}
