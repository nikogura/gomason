package gomason

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// PublishFile publishes the binary to wherever you have it configured to go
func PublishFile(meta Metadata, filePath string, verbose bool) (err error) {
	// get creds
	username, password, err := GetCredentials(meta, verbose)
	if err != nil {
		err = errors.Wrapf(err, "failed to get credentials")
		return err
	}

	if verbose {
		log.Printf("Publishing %s", filePath)
	}

	client := &http.Client{}

	fileName := filepath.Base(filePath)

	target, ok := meta.PublishInfo.TargetsMap[fileName]

	if ok {
		// upload the file
		err = UploadFile(client, target.Destination, filePath, meta, username, password, verbose)
		if err != nil {
			err = errors.Wrapf(err, "failed to upload file %s", filePath)
			return err
		}

		// upload the detached signature
		if target.Signature {
			err := UploadSignature(client, target.Destination, filePath, meta, username, password, verbose)
			if err != nil {
				err = errors.Wrapf(err, "failed to upload signature for %s", filePath)
				return err
			}

		}

		// upload the checksums if configured to do so
		if target.Checksums {
			err := UploadChecksums(client, target.Destination, filePath, meta, username, password, verbose)
			if err != nil {
				err = errors.Wrapf(err, "failed to upload checksums for %s", filePath)
				return err
			}
		}
	}

	return err
}

// UploadChecksums uploads the checksums for a file.  This is useful if the repository is not configured to do so automatically.
func UploadChecksums(client *http.Client, destination, filename string, meta Metadata, username string, password string, verbose bool) (err error) {

	md5sum, sha1sum, sha256sum, err := AllChecksumsForFile(filename)
	if err != nil {
		err = errors.Wrapf(err, "failed to calculate checksum for %s", filename)
		return err
	}
	// write each checksum file
	// checksum format: <sum> <filename>\n

	base := filepath.Base(filename)

	parsedDestination, err := ParseTemplateForMetadata(destination, meta)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse destination url %s", destination)
		return err
	}

	// upload Md5Sum
	err = UploadChecksum(parsedDestination, md5sum, "md5", base, client, username, password, verbose)
	if err != nil {
		err = errors.Wrapf(err, "failed to upload md5sum file for %s", filename)
	}

	// upload Sha1Sum
	err = UploadChecksum(parsedDestination, sha1sum, "sha1", base, client, username, password, verbose)
	if err != nil {
		err = errors.Wrapf(err, "failed to upload sha1sum file for %s", filename)
	}

	// upload Sha256Sum
	err = UploadChecksum(parsedDestination, sha256sum, "sha256", base, client, username, password, verbose)
	if err != nil {
		err = errors.Wrapf(err, "failed to upload sha256sum file for %s", filename)
	}

	return err
}

// UploadChecksum uploads the checksum of the given type for the given file
func UploadChecksum(parsedDestination, checksum, sumtype, fileName string, client *http.Client, username, password string, verbose bool) (err error) {
	// upload Md5 sum
	target := fmt.Sprintf("%s.%s", parsedDestination, sumtype)
	contents := fmt.Sprintf("%s %s\n", checksum, fileName)

	sumMd5, sumSha1, sumSha256, err := AllChecksumsForBytes([]byte(contents))
	if err != nil {
		err = errors.Wrapf(err, "failed to generate checksums for %s file with contents %q", sumtype, contents)
		return err
	}

	if verbose {
		log.Printf("Uploading checksum to %s", target)
	}

	err = Upload(client, target, strings.NewReader(contents), sumMd5, sumSha1, sumSha256, username, password, verbose)
	if err != nil {
		err = errors.Wrapf(err, "failed to upload md5sum file to %s", target)
		return err
	}

	return err
}

// UploadFile uploads a file off the filesystem
func UploadFile(client *http.Client, destination string, filename string, meta Metadata, username string, password string, verbose bool) (err error) {
	// get data
	data, err := os.Open(filename)
	if err != nil {
		err = errors.Wrapf(err, "failed to open %s", filename)
		return err
	}

	// get checksums
	md5sum, sha1sum, sha256sum, err := AllChecksumsForFile(filename)
	if err != nil {
		err = errors.Wrapf(err, "failed to calculate checksum for %s", filename)
		return err
	}

	parsedDestination, err := ParseTemplateForMetadata(destination, meta)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse destination url %s", destination)
		return err
	}

	if verbose {
		log.Printf("Attempting to upload %s to %s", filename, parsedDestination)
	}

	return Upload(client, parsedDestination, data, md5sum, sha1sum, sha256sum, username, password, verbose)

}

// UploadSignature uploads the detached signature for a file
func UploadSignature(client *http.Client, destination string, filename string, meta Metadata, username string, password string, verbose bool) (err error) {
	filename += ".asc"
	destination += ".asc"

	return UploadFile(client, destination, filename, meta, username, password, verbose)
}

// Upload actually does the upload.  It uploads pure data.
func Upload(client *http.Client, parsedDestination string, data io.Reader, md5sum string, sha1sum string, sha256sum string, username string, password string, verbose bool) (err error) {

	req, err := http.NewRequest("PUT", parsedDestination, data)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to create http request for target %s", parsedDestination))
		return err
	}

	// add headers  (Technically these are what Artifactory expects, but should be fine for any REST interface)
	req.Header.Add("X-Checksum-Md5", md5sum)
	req.Header.Add("X-Checksum-Sha1", sha1sum)
	req.Header.Add("X-Checksum-Sha256", sha256sum)
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("Failed to PUT to url %s", parsedDestination))
		return err
	}

	if verbose {
		log.Printf("Response: %s", resp.Status)
		log.Printf("Response Code: %d", resp.StatusCode)
	}

	if resp.StatusCode > 299 {
		err = errors.Wrap(err, "response code %d is not indicative of a successful publish")
		return err
	}

	return err
}
