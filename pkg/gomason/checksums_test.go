package gomason

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytesMd5(t *testing.T) {
	md5sum, err := BytesMd5([]byte(testFileContent()))
	if err != nil {
		log.Printf("Failed to md5sum test file: %s", err)
		t.Fail()
	}

	log.Printf("md5sum: %s", md5sum)

	assert.Equal(t, testFileMd5(), md5sum, "Generated md5sum matches expectations.")
}

func TestFileMd5(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "checksum")
	if err != nil {
		log.Fatal("Failed to create temp dir for testing")
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := fmt.Sprintf("%s/testfile", tmpDir)
	err = ioutil.WriteFile(tmpFile, []byte(testFileContent()), 0644)
	if err != nil {
		log.Printf("Failed to write temp file: %s", err)
		t.Fail()
	}

	md5sum, err := FileMd5(tmpFile)
	if err != nil {
		log.Printf("Failed to md5sum test file: %s", err)
		t.Fail()
	}

	log.Printf("md5sum: %s", md5sum)

	assert.Equal(t, testFileMd5(), md5sum, "Generated md5sum matches expectations.")
}

func TestBytesSha1(t *testing.T) {
	sha1sum, err := BytesSha1([]byte(testFileContent()))
	if err != nil {
		log.Printf("Failed to md5sum test file: %s", err)
		t.Fail()
	}

	log.Printf("Sha1sum: %s", sha1sum)

	assert.Equal(t, testFileSha1(), sha1sum, "Generated sha1sum matches expectations.")
}

func TestFileSha1(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "checksum")
	if err != nil {
		log.Fatal("Failed to create temp dir for testing")
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := fmt.Sprintf("%s/testfile", tmpDir)
	err = ioutil.WriteFile(tmpFile, []byte(testFileContent()), 0644)
	if err != nil {
		log.Printf("Failed to write temp file: %s", err)
		t.Fail()
	}

	sha1sum, err := FileSha1(tmpFile)
	if err != nil {
		log.Printf("Failed to md5sum test file: %s", err)
		t.Fail()
	}

	log.Printf("Sha1sum: %s", sha1sum)

	assert.Equal(t, testFileSha1(), sha1sum, "Generated sha1sum matches expectations.")
}

func TestBytesSha256(t *testing.T) {
	sha256sum, err := BytesSha256([]byte(testFileContent()))

	if err != nil {
		log.Printf("Failed to md5sum test file: %s", err)
		t.Fail()
	}

	log.Printf("Sha256sum: %s", sha256sum)

	assert.Equal(t, testFileSha256(), sha256sum, "Generated sha1sum matches expectations.")
}

func TestFileSha256(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "checksum")
	if err != nil {
		log.Fatal("Failed to create temp dir for testing")
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := fmt.Sprintf("%s/testfile", tmpDir)
	err = ioutil.WriteFile(tmpFile, []byte(testFileContent()), 0644)
	if err != nil {
		log.Printf("Failed to write temp file: %s", err)
		t.Fail()
	}

	sha256sum, err := FileSha256(tmpFile)
	if err != nil {
		log.Printf("Failed to md5sum test file: %s", err)
		t.Fail()
	}

	log.Printf("Sha256sum: %s", sha256sum)

	assert.Equal(t, testFileSha256(), sha256sum, "Generated sha1sum matches expectations.")
}

func TestAllChecksumsForBytes(t *testing.T) {
	md5sum, sha1sum, sha256sum, err := AllChecksumsForBytes([]byte(testFileContent()))
	if err != nil {
		log.Printf("Failed to generate checksums for byte array")
		t.Fail()
	}

	expected := testAllChecksums()

	actual := []string{md5sum, sha1sum, sha256sum}

	assert.Equal(t, expected, actual, "Returned checksum list meets expectations")

}
