package gomason

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestBytesMd5(t *testing.T) {
	md5sum, err := BytesMd5([]byte(testFileContent()))
	if err != nil {
		t.Errorf("Failed to md5sum test file: %s\n", err)
	}

	assert.Equal(t, testFileMd5(), md5sum, "Generated md5sum matches expectations.")
}

func TestFileMd5(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "checksum")
	if err != nil {
		t.Errorf("Failed to create temp dir for testing: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := fmt.Sprintf("%s/testfile", tmpDir)
	err = os.WriteFile(tmpFile, []byte(testFileContent()), 0644)
	if err != nil {
		t.Errorf("Failed to write temp file: %s", err)
	}

	md5sum, err := FileMd5(tmpFile)
	if err != nil {
		t.Errorf("Failed to md5sum test file: %s", err)
	}

	assert.Equal(t, testFileMd5(), md5sum, "Generated md5sum matches expectations.")
}

func TestBytesSha1(t *testing.T) {
	sha1sum, err := BytesSha1([]byte(testFileContent()))
	if err != nil {
		t.Errorf("Failed to md5sum test file: %s", err)
	}

	assert.Equal(t, testFileSha1(), sha1sum, "Generated sha1sum matches expectations.")
}

func TestFileSha1(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "checksum")
	if err != nil {
		t.Errorf("Failed to create temp dir for testing: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := fmt.Sprintf("%s/testfile", tmpDir)
	err = os.WriteFile(tmpFile, []byte(testFileContent()), 0644)
	if err != nil {
		t.Errorf("Failed to write temp file: %s", err)
	}

	sha1sum, err := FileSha1(tmpFile)
	if err != nil {
		t.Errorf("Failed to md5sum test file: %s", err)
	}

	assert.Equal(t, testFileSha1(), sha1sum, "Generated sha1sum matches expectations.")
}

func TestBytesSha256(t *testing.T) {
	sha256sum, err := BytesSha256([]byte(testFileContent()))

	if err != nil {
		t.Errorf("Failed to md5sum test file: %s", err)
	}

	assert.Equal(t, testFileSha256(), sha256sum, "Generated sha1sum matches expectations.")
}

func TestFileSha256(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "checksum")
	if err != nil {
		t.Errorf("Failed to create temp dir for testing: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := fmt.Sprintf("%s/testfile", tmpDir)
	err = os.WriteFile(tmpFile, []byte(testFileContent()), 0644)
	if err != nil {
		t.Errorf("Failed to write temp file: %s", err)
	}

	sha256sum, err := FileSha256(tmpFile)
	if err != nil {
		t.Errorf("Failed to md5sum test file: %s", err)
	}

	assert.Equal(t, testFileSha256(), sha256sum, "Generated sha1sum matches expectations.")
}

func TestAllChecksumsForBytes(t *testing.T) {
	md5sum, sha1sum, sha256sum, err := AllChecksumsForBytes([]byte(testFileContent()))
	if err != nil {
		t.Errorf("Failed to generate checksums for byte array")
	}

	expected := testAllChecksums()

	actual := []string{md5sum, sha1sum, sha256sum}

	assert.Equal(t, expected, actual, "Returned checksum list meets expectations")

}
