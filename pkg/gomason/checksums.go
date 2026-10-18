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
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"os"

	"github.com/pkg/errors"
)

// BytesMd5 generates the md5sum for a byte array.
func BytesMd5(input []byte) (checksum string, err error) {
	hasher := md5.New()

	_, err = hasher.Write(input)
	if err != nil {
		return checksum, err
	}

	checksum = hex.EncodeToString(hasher.Sum(nil))

	return checksum, err
}

// FileMd5 generates the md5sum for a file.
func FileMd5(filename string) (checksum string, err error) {
	var checksumBytes []byte

	checksumBytes, err = os.ReadFile(filename)
	if err != nil {
		return checksum, err
	}

	checksum, err = BytesMd5(checksumBytes)

	return checksum, err
}

// BytesSha1 generates the sha1sum for a byte array.
func BytesSha1(input []byte) (checksum string, err error) {
	hasher := sha1.New()

	_, err = hasher.Write(input)
	if err != nil {
		return checksum, err
	}

	checksum = hex.EncodeToString(hasher.Sum(nil))

	return checksum, err

}

// FileSha1 generates the sha1sum for a file.
func FileSha1(filename string) (checksum string, err error) {
	var checksumBytes []byte

	checksumBytes, err = os.ReadFile(filename)
	if err != nil {
		return checksum, err
	}

	checksum, err = BytesSha1(checksumBytes)

	return checksum, err
}

// BytesSha256 generates the sha256sum for a byte array.
func BytesSha256(input []byte) (checksum string, err error) {
	hasher := sha256.New()

	_, err = hasher.Write(input)
	if err != nil {
		return checksum, err
	}

	checksum = hex.EncodeToString(hasher.Sum(nil))

	return checksum, err
}

// FileSha256 generates the Sha256 sum for a file.
func FileSha256(filename string) (checksum string, err error) {
	var checksumBytes []byte

	checksumBytes, err = os.ReadFile(filename)
	if err != nil {
		return checksum, err
	}

	checksum, err = BytesSha256(checksumBytes)

	return checksum, err
}

// AllChecksumsForFile is a convenience method that generates and returns md5, sha1, and sha256 checksums for a given file.
func AllChecksumsForFile(filename string) (md5sum, sha1sum, sha256sum string, err error) {
	md5sum, err = FileMd5(filename)
	if err != nil {
		err = errors.Wrapf(err, "failed to calculate md5sum for %s", filename)
		return md5sum, sha1sum, sha256sum, err
	}

	sha1sum, err = FileSha1(filename)
	if err != nil {
		err = errors.Wrapf(err, "failed to calculate sha1sum for %s", filename)
		return md5sum, sha1sum, sha256sum, err
	}

	sha256sum, err = FileSha256(filename)
	if err != nil {
		err = errors.Wrapf(err, "failed to calculate sha256sum for %s", filename)
		return md5sum, sha1sum, sha256sum, err
	}

	return md5sum, sha1sum, sha256sum, err
}

// AllChecksumsForBytes is a convenience method for returning the md5, sha1, sha256 checksums for a byte array.
func AllChecksumsForBytes(input []byte) (md5sum, sha1sum, sha256sum string, err error) {
	md5sum, err = BytesMd5(input)
	if err != nil {
		err = errors.Wrapf(err, "failed to calculate md5sum for %s", string(input))
		return md5sum, sha1sum, sha256sum, err
	}

	sha1sum, err = BytesSha1(input)
	if err != nil {
		err = errors.Wrapf(err, "failed to calculate sha1sum for %s", string(input))
		return md5sum, sha1sum, sha256sum, err
	}

	sha256sum, err = BytesSha256(input)
	if err != nil {
		err = errors.Wrapf(err, "failed to calculate sha256sum for %s", string(input))
		return md5sum, sha1sum, sha256sum, err
	}

	return md5sum, sha1sum, sha256sum, err
}
