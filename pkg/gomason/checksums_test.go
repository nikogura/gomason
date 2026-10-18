// Copyright © 2017 Nik Ogura <nik.ogura@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gomason

import (
	"fmt"
	"os"
	"testing"

	"github.com/nikogura/gomason/v2/pkg/gomason/testfixtures"
	"github.com/stretchr/testify/assert"
)

func TestBytesMd5(t *testing.T) {
	md5sum, err := BytesMd5([]byte(testfixtures.FileContent))
	if err != nil {
		t.Errorf("Failed to md5sum test file: %s\n", err)
	}

	assert.Equal(t, testfixtures.FileMd5, md5sum, "Generated md5sum matches expectations.")
}

func TestFileMd5(t *testing.T) {
	tmpDir := t.TempDir()

	tmpFile := fmt.Sprintf("%s/testfile", tmpDir)
	err := os.WriteFile(tmpFile, []byte(testfixtures.FileContent), 0644)
	if err != nil {
		t.Errorf("Failed to write temp file: %s", err)
	}

	md5sum, err := FileMd5(tmpFile)
	if err != nil {
		t.Errorf("Failed to md5sum test file: %s", err)
	}

	assert.Equal(t, testfixtures.FileMd5, md5sum, "Generated md5sum matches expectations.")
}

func TestBytesSha1(t *testing.T) {
	sha1sum, err := BytesSha1([]byte(testfixtures.FileContent))
	if err != nil {
		t.Errorf("Failed to md5sum test file: %s", err)
	}

	assert.Equal(t, testfixtures.FileSha1, sha1sum, "Generated sha1sum matches expectations.")
}

func TestFileSha1(t *testing.T) {
	tmpDir := t.TempDir()

	tmpFile := fmt.Sprintf("%s/testfile", tmpDir)
	err := os.WriteFile(tmpFile, []byte(testfixtures.FileContent), 0644)
	if err != nil {
		t.Errorf("Failed to write temp file: %s", err)
	}

	sha1sum, err := FileSha1(tmpFile)
	if err != nil {
		t.Errorf("Failed to md5sum test file: %s", err)
	}

	assert.Equal(t, testfixtures.FileSha1, sha1sum, "Generated sha1sum matches expectations.")
}

func TestBytesSha256(t *testing.T) {
	sha256sum, err := BytesSha256([]byte(testfixtures.FileContent))

	if err != nil {
		t.Errorf("Failed to md5sum test file: %s", err)
	}

	assert.Equal(t, testfixtures.FileSha256, sha256sum, "Generated sha1sum matches expectations.")
}

func TestFileSha256(t *testing.T) {
	tmpDir := t.TempDir()

	tmpFile := fmt.Sprintf("%s/testfile", tmpDir)
	err := os.WriteFile(tmpFile, []byte(testfixtures.FileContent), 0644)
	if err != nil {
		t.Errorf("Failed to write temp file: %s", err)
	}

	sha256sum, err := FileSha256(tmpFile)
	if err != nil {
		t.Errorf("Failed to md5sum test file: %s", err)
	}

	assert.Equal(t, testfixtures.FileSha256, sha256sum, "Generated sha1sum matches expectations.")
}

func TestAllChecksumsForBytes(t *testing.T) {
	md5sum, sha1sum, sha256sum, err := AllChecksumsForBytes([]byte(testfixtures.FileContent))
	if err != nil {
		t.Errorf("Failed to generate checksums for byte array")
	}

	expected := testfixtures.AllChecksums()

	actual := []string{md5sum, sha1sum, sha256sum}

	assert.Equal(t, expected, actual, "Returned checksum list meets expectations")

}
