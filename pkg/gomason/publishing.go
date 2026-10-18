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
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"                  //nolint:staticcheck // aws-sdk-go v1 is past end-of-support; migrating the S3 publish path to aws-sdk-go-v2 is tracked separately.
	"github.com/aws/aws-sdk-go/aws/session"          //nolint:staticcheck // aws-sdk-go v1 is past end-of-support; migrating the S3 publish path to aws-sdk-go-v2 is tracked separately.
	"github.com/aws/aws-sdk-go/service/s3"           //nolint:staticcheck // aws-sdk-go v1 is past end-of-support; migrating the S3 publish path to aws-sdk-go-v2 is tracked separately.
	"github.com/aws/aws-sdk-go/service/s3/s3manager" //nolint:staticcheck // aws-sdk-go v1 is past end-of-support; migrating the S3 publish path to aws-sdk-go-v2 is tracked separately.
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// PublishFile publishes the binary to wherever you have it configured to go.
func (g *Gomason) PublishFile(meta Metadata, filePath string) (err error) {
	// get creds
	var username, password, bearerToken string

	username, password, bearerToken, err = g.GetCredentials(meta)
	if err != nil {
		err = errors.Wrapf(err, "failed to get credentials")
		return err
	}

	logrus.Debugf("Publishing %s", filePath)

	client := &http.Client{}

	fileName := filepath.Base(filePath)

	target, ok := meta.PublishInfo.TargetsMap[fileName]
	if !ok {
		return err
	}

	// upload the file
	err = UploadFile(client, target.Destination, filePath, meta, username, password, bearerToken)
	if err != nil {
		err = errors.Wrapf(err, "failed to upload file %s", filePath)
		return err
	}

	// upload the detached signature
	if target.Signature {
		err = UploadSignature(client, target.Destination, filePath, meta, username, password, bearerToken)
		if err != nil {
			err = errors.Wrapf(err, "failed to upload signature for %s", filePath)
			return err
		}
	}

	// upload the checksums if configured to do so
	if target.Checksums {
		err = UploadChecksums(client, target.Destination, filePath, meta, username, password, bearerToken)
		if err != nil {
			err = errors.Wrapf(err, "failed to upload checksums for %s", filePath)
			return err
		}
	}

	// upload the attestation files (provenance, SBOM, and their signatures) if configured
	// to do so. Fail closed: the target asks for attestations but the files were never
	// generated — that mismatch is an error, not a silent partial publish.
	if target.Attestations {
		err = UploadAttestationFiles(client, target.Destination, filePath, meta, username, password, bearerToken)
		if err != nil {
			err = errors.Wrapf(err, "failed to upload attestations for %s", filePath)
			return err
		}
	}

	return err
}

// UploadAttestationFiles uploads the provenance attestation, the SBOM attestation, and
// each one's detached signature, beside the artifact's destination — the file layout a
// distribution client (e.g. dbt) expects: <destination>.att.json, .att.json.asc,
// .sbom.json, .sbom.json.asc.
func UploadAttestationFiles(client *http.Client, destination string, filename string, meta Metadata, username string, password string, bearerToken string) (err error) {
	for _, suffix := range []string{".att.json", ".att.json.asc", ".sbom.json", ".sbom.json.asc"} {
		attFile := fmt.Sprintf("%s%s", filename, suffix)

		_, statErr := os.Stat(attFile)
		if os.IsNotExist(statErr) {
			err = fmt.Errorf("attestation file %s is missing: generate attestations before publishing (building.attestations in metadata, and syft on PATH)", attFile)
			return err
		}

		if statErr != nil {
			err = errors.Wrapf(statErr, "failed to stat %s", attFile)
			return err
		}

		parsedDestination, parseErr := ParseTemplateForMetadata(fmt.Sprintf("%s%s", destination, suffix), meta)
		if parseErr != nil {
			err = errors.Wrapf(parseErr, "failed to parse destination url %s", destination)
			return err
		}

		err = UploadFile(client, parsedDestination, attFile, meta, username, password, bearerToken)
		if err != nil {
			err = errors.Wrapf(err, "failed to upload %s", attFile)
			return err
		}
	}

	return err
}

// UploadChecksums uploads the checksums for a file.  This is useful if the repository is not configured to do so automatically.
func UploadChecksums(client *http.Client, destination, filename string, meta Metadata, username string, password string, bearerToken string) (err error) {

	var md5sum, sha1sum, sha256sum string

	md5sum, sha1sum, sha256sum, err = AllChecksumsForFile(filename)
	if err != nil {
		err = errors.Wrapf(err, "failed to calculate checksum for %s", filename)
		return err
	}

	var parsedDestination string

	parsedDestination, err = ParseTemplateForMetadata(destination, meta)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse destination url %s", destination)
		return err
	}

	// upload Md5Sum
	err = UploadChecksum(parsedDestination, md5sum, "md5", client, username, password, bearerToken)
	if err != nil {
		err = errors.Wrapf(err, "failed to upload md5sum file for %s", filename)
		return err
	}

	// upload Sha1Sum
	err = UploadChecksum(parsedDestination, sha1sum, "sha1", client, username, password, bearerToken)
	if err != nil {
		err = errors.Wrapf(err, "failed to upload sha1sum file for %s", filename)
		return err
	}

	// upload Sha256Sum
	err = UploadChecksum(parsedDestination, sha256sum, "sha256", client, username, password, bearerToken)
	if err != nil {
		err = errors.Wrapf(err, "failed to upload sha256sum file for %s", filename)
	}

	return err
}

// UploadChecksum uploads the checksum of the given type for the given file.
func UploadChecksum(parsedDestination, checksum, sumtype string, client *http.Client, username, password, bearerToken string) (err error) {
	target := fmt.Sprintf("%s.%s", parsedDestination, sumtype)
	contents := checksum

	var sumMd5, sumSha1, sumSha256 string

	sumMd5, sumSha1, sumSha256, err = AllChecksumsForBytes([]byte(contents))
	if err != nil {
		err = errors.Wrapf(err, "failed to generate checksums for %s file with contents %q", sumtype, contents)
		return err
	}

	logrus.Debugf("Uploading checksum to %s", target)

	err = Upload(client, target, strings.NewReader(contents), sumMd5, sumSha1, sumSha256, username, password, bearerToken)
	if err != nil {
		err = errors.Wrapf(err, "failed to upload md5sum file to %s", target)
		return err
	}

	return err
}

// UploadFile uploads a file off the filesystem.
func UploadFile(client *http.Client, destination string, filename string, meta Metadata, username string, password string, bearerToken string) (err error) {
	// get data
	var data *os.File

	data, err = os.Open(filename)
	if err != nil {
		err = errors.Wrapf(err, "failed to open %s", filename)
		return err
	}

	// get checksums
	var md5sum, sha1sum, sha256sum string

	md5sum, sha1sum, sha256sum, err = AllChecksumsForFile(filename)
	if err != nil {
		err = errors.Wrapf(err, "failed to calculate checksum for %s", filename)
		return err
	}

	var parsedDestination string

	parsedDestination, err = ParseTemplateForMetadata(destination, meta)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse destination url %s", destination)
		return err
	}

	logrus.Debugf("Attempting to upload %s to %s", filename, parsedDestination)

	err = Upload(client, parsedDestination, data, md5sum, sha1sum, sha256sum, username, password, bearerToken)

	return err
}

// UploadSignature uploads the detached signature for a file.
func UploadSignature(client *http.Client, destination string, filename string, meta Metadata, username string, password string, bearerToken string) (err error) {
	filename += ".asc"
	destination += ".asc"

	err = UploadFile(client, destination, filename, meta, username, password, bearerToken)

	return err
}

// Upload actually does the upload.  It uploads pure data.  If bearerToken is non-empty, it is used as an Authorization: Bearer header instead of basic auth.
func Upload(client *http.Client, url string, data io.Reader, md5sum string, sha1sum string, sha256sum string, username string, password string, bearerToken string) (err error) {
	// Check to see if this is an S3 URL
	isS3, s3Meta := S3Url(url)

	if isS3 {
		err = uploadToS3(url, data, s3Meta)
		return err
	}

	err = uploadToHTTP(client, url, data, md5sum, sha1sum, sha256sum, username, password, bearerToken)

	return err
}

// uploadToS3 handles uploading data to an S3 bucket.
func uploadToS3(url string, data io.Reader, s3Meta S3Meta) (err error) {
	var sess *session.Session

	sess, err = DefaultSession()
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

	err = createS3Directories(sess, s3Meta)

	return err
}

// createS3Directories creates the directory structure (0-byte objects) in S3.
func createS3Directories(sess *session.Session, s3Meta S3Meta) (err error) {
	var dirs []string

	dirs, err = DirsForURL(s3Meta.Key)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse dirs for %s", s3Meta.Key)
		return err
	}

	headSvc := s3.New(sess)
	s3Client := s3.New(sess)

	// create the 'folders' (0 byte objects) in s3
	for _, d := range dirs {
		if d == "." {
			continue
		}

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

	return err
}

// uploadToHTTP handles uploading data via HTTP PUT.
func uploadToHTTP(client *http.Client, url string, data io.Reader, md5sum string, sha1sum string, sha256sum string, username string, password string, bearerToken string) (err error) {
	var req *http.Request

	req, err = http.NewRequestWithContext(context.Background(), http.MethodPut, url, data)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to create http request for target %s", url))
		return err
	}

	// add headers  (Technically these are what Artifactory expects, but should be fine for any REST interface)
	req.Header.Add("X-Checksum-Md5", md5sum)
	req.Header.Add("X-Checksum-Sha1", sha1sum)
	req.Header.Add("X-Checksum-Sha256", sha256sum)

	// Bearer token takes precedence over basic auth
	if bearerToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearerToken))
	} else {
		req.SetBasicAuth(username, password)
	}

	var resp *http.Response

	resp, err = client.Do(req)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("Failed to PUT to url %s", url))
		return err
	}
	defer resp.Body.Close()

	logrus.Debugf("Response: %s", resp.Status)
	logrus.Debugf("Response Code: %d", resp.StatusCode)

	if resp.StatusCode > 299 {
		err = errors.New(fmt.Sprintf("response code %d is not indicative of a successful publish", resp.StatusCode))
		return err
	}

	return err
}
