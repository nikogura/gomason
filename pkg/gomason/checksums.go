package gomason

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"github.com/pkg/errors"
	"io/ioutil"
)

// BytesMd5 generates the md5sum for a byte array
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
	checksumBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return checksum, err
	}

	return BytesMd5(checksumBytes)
}

//BytesSha1 generates the sha1sum for a byte array
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
	checksumBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return checksum, err
	}

	return BytesSha1(checksumBytes)
}

// BytesSha256 generates the sha256sum for a byte array
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
	checksumBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return checksum, err
	}

	return BytesSha256(checksumBytes)
}

// AllChecksumsForFile is a convenience method that generates and returns md5, sha1, and sha256 checksums for a given file
func AllChecksumsForFile(filename string) (md5sum, sha1sum, sha256sum string, err error) {
	md5sum, err = FileMd5(filename)
	if err != nil {
		err = errors.Wrapf(err, "failed to calculate md5sum for %s", filename)
		return
	}

	sha1sum, err = FileSha1(filename)
	if err != nil {
		err = errors.Wrapf(err, "failed to calculate sha1sum for %s", filename)
		return
	}

	sha256sum, err = FileSha256(filename)
	if err != nil {
		err = errors.Wrapf(err, "failed to calculate sha256sum for %s", filename)
		return
	}

	return
}

// AllChecksumsForBytes is a convenience method for returning the md5, sha1, sha256 checksums for a byte array.
func AllChecksumsForBytes(input []byte) (md5sum, sha1sum, sha256sum string, err error) {
	md5sum, err = BytesMd5(input)
	if err != nil {
		err = errors.Wrapf(err, "failed to calculate md5sum for %s", string(input))
		return
	}

	sha1sum, err = BytesSha1(input)
	if err != nil {
		err = errors.Wrapf(err, "failed to calculate sha1sum for %s", string(input))
		return
	}

	sha256sum, err = BytesSha256(input)
	if err != nil {
		err = errors.Wrapf(err, "failed to calculate sha256sum for %s", string(input))
		return
	}

	return
}
