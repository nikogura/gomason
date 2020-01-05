package gomason

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testFileContent() string {
	return `the quick fox jumped over the lazy brown dog`
}

func testFileMd5() string {
	return "356b5768c6964531f678781446840b76"
}

func testFileSha1() string {
	return "041b2390cd9697ba6b9f57b532b0aa5ac183736b"
}

func testFileSha256() string {
	return "e088f8b9456b8a91a48159497ac425a4c3cdcad3ad81cc3a269618209dee033b"
}

func testRawUrl() string {
	return "http://localhost:8081/artifactory/repo-local/foo/{{.Version}}/linux/amd64/foo"
}

func testParsedUrl(version string) string {
	return fmt.Sprintf("http://localhost:8081/artifactory/repo-local/foo/%s/linux/amd64/foo", version)
}

func testAllChecksums() []string {
	return []string{testFileMd5(), testFileSha1(), testFileSha256()}
}

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
