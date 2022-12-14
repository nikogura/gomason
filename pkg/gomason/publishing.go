package gomason

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// PublishFile publishes the binary to wherever you have it configured to go
func (g *Gomason) PublishFile(meta Metadata, filePath string) (err error) {
	// get creds
	username, password, err := g.GetCredentials(meta)
	if err != nil {
		err = errors.Wrapf(err, "failed to get credentials")
		return err
	}

	logrus.Debugf("Publishing %s", filePath)

	client := &http.Client{}

	fileName := filepath.Base(filePath)

	target, ok := meta.PublishInfo.TargetsMap[fileName]

	if ok {
		// upload the file
		err = UploadFile(client, target.Destination, filePath, meta, username, password)
		if err != nil {
			err = errors.Wrapf(err, "failed to upload file %s", filePath)
			return err
		}

		// upload the detached signature
		if target.Signature {
			err := UploadSignature(client, target.Destination, filePath, meta, username, password)
			if err != nil {
				err = errors.Wrapf(err, "failed to upload signature for %s", filePath)
				return err
			}

		}

		// upload the checksums if configured to do so
		if target.Checksums {
			err := UploadChecksums(client, target.Destination, filePath, meta, username, password)
			if err != nil {
				err = errors.Wrapf(err, "failed to upload checksums for %s", filePath)
				return err
			}
		}
	}

	return err
}

// UploadChecksums uploads the checksums for a file.  This is useful if the repository is not configured to do so automatically.
func UploadChecksums(client *http.Client, destination, filename string, meta Metadata, username string, password string) (err error) {

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

	// upload Md5Sum
	err = UploadChecksum(parsedDestination, md5sum, "md5", client, username, password)
	if err != nil {
		err = errors.Wrapf(err, "failed to upload md5sum file for %s", filename)
	}

	// upload Sha1Sum
	err = UploadChecksum(parsedDestination, sha1sum, "sha1", client, username, password)
	if err != nil {
		err = errors.Wrapf(err, "failed to upload sha1sum file for %s", filename)
	}

	// upload Sha256Sum
	err = UploadChecksum(parsedDestination, sha256sum, "sha256", client, username, password)
	if err != nil {
		err = errors.Wrapf(err, "failed to upload sha256sum file for %s", filename)
	}

	return err
}

// UploadChecksum uploads the checksum of the given type for the given file
func UploadChecksum(parsedDestination, checksum, sumtype string, client *http.Client, username, password string) (err error) {
	target := fmt.Sprintf("%s.%s", parsedDestination, sumtype)
	contents := checksum

	sumMd5, sumSha1, sumSha256, err := AllChecksumsForBytes([]byte(contents))
	if err != nil {
		err = errors.Wrapf(err, "failed to generate checksums for %s file with contents %q", sumtype, contents)
		return err
	}

	logrus.Debugf("Uploading checksum to %s", target)

	err = Upload(client, target, strings.NewReader(contents), sumMd5, sumSha1, sumSha256, username, password)
	if err != nil {
		err = errors.Wrapf(err, "failed to upload md5sum file to %s", target)
		return err
	}

	return err
}

// UploadFile uploads a file off the filesystem
func UploadFile(client *http.Client, destination string, filename string, meta Metadata, username string, password string) (err error) {
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

	logrus.Debugf("Attempting to upload %s to %s", filename, parsedDestination)

	return Upload(client, parsedDestination, data, md5sum, sha1sum, sha256sum, username, password)
}

// UploadSignature uploads the detached signature for a file
func UploadSignature(client *http.Client, destination string, filename string, meta Metadata, username string, password string) (err error) {
	filename += ".asc"
	destination += ".asc"

	return UploadFile(client, destination, filename, meta, username, password)
}

// Upload actually does the upload.  It uploads pure data.
func Upload(client *http.Client, url string, data io.Reader, md5sum string, sha1sum string, sha256sum string, username string, password string) (err error) {
	// Check to see if this is an S3 URL
	isS3, s3Meta := S3Url(url)

	if isS3 {
		sess, err := DefaultSession()
		if err != nil {
			err = errors.Wrap(err, "Failed to create AWS session")
			return err
		}

		uploader := s3manager.NewUploader(sess)

		uploadOptions := &s3manager.UploadInput{
			Body:   data,
			Bucket: aws.String(s3Meta.Bucket),
			Key:    aws.String(s3Meta.Key),
		}

		_, err = uploader.Upload(uploadOptions)
		if err != nil {
			err = errors.Wrapf(err, "failed uploading to %s", url)
			return err
		}

		// make the directory paths in s3
		dirs, err := DirsForURL(s3Meta.Key)
		if err != nil {
			err = errors.Wrapf(err, "failed to parse dirs for %s", s3Meta.Key)
			return err
		}

		headSvc := s3.New(sess)
		s3Client := s3.New(sess)

		// create the 'folders' (0 byte objects) in s3
		for _, d := range dirs {
			if d != "." {
				path := fmt.Sprintf("%s/", d)
				// check to see if it doesn't already exist
				headOptions := &s3.HeadObjectInput{
					Bucket: aws.String(s3Meta.Bucket),
					Key:    aws.String(path),
				}

				_, err = headSvc.HeadObject(headOptions)
				// if there's an error, it doesn't exist
				if err != nil {
					// so create it
					_, err = s3Client.PutObject(&s3.PutObjectInput{
						Bucket: aws.String(s3Meta.Bucket),
						Key:    aws.String(path),
					})
					if err != nil {
						err = errors.Wrapf(err, "Failed to create %s in s3", path)
						return err
					}
				}
			}
		}

		return err

	}

	// TODO Check if destination exists
	// TODO Create path if not
	req, err := http.NewRequest("PUT", url, data)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to create http request for target %s", url))
		return err
	}

	// add headers  (Technically these are what Artifactory expects, but should be fine for any REST interface)
	req.Header.Add("X-Checksum-Md5", md5sum)
	req.Header.Add("X-Checksum-Sha1", sha1sum)
	req.Header.Add("X-Checksum-Sha256", sha256sum)
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("Failed to PUT to url %s", url))
		return err
	}

	logrus.Debugf("Response: %s", resp.Status)
	logrus.Debugf("Response Code: %d", resp.StatusCode)

	if resp.StatusCode > 299 {
		err = errors.New(fmt.Sprintf("response code %d is not indicative of a successful publish", resp.StatusCode))
		return err
	}

	return err
}
