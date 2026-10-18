// Copyright © 2026 Nik Ogura <nik.ogura@gmail.com>
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

// Package testfixtures provides static test data for gomason tests.
//
// The fixtures here are deliberately self-contained: this package imports
// nothing from gomason, so the gomason package's own in-package tests can
// consume it without an import cycle.  Anything that needs a gomason type
// belongs in a _test.go file in the gomason package, not here.
package testfixtures

// ModuleName is the go package name of the project the tests build against.
const ModuleName = "github.com/nikogura/testproject"

// Version is the version carried by the test metadata.
const Version = "0.1.0"

// Description is the description carried by the test metadata.
const Description = "Test Project for Gomason."

// SigningProgram is the signing program named in the test metadata.
const SigningProgram = "gpg"

// SigningEmail is the signing identity named in the test metadata.
const SigningEmail = "gomason-tester@foo.com"

// PrepCommand is the single prep command run by the test metadata.
const PrepCommand = `echo "GOPATH is: ${GOPATH}"`

// BuildTargetName is the build target named in the test metadata.
const BuildTargetName = "linux/amd64"

// PublishSource is the artifact name produced by a test build.  It is the
// module's base name joined to the build target, per gomason's naming.
const PublishSource = "testproject_linux_amd64"

// PublishDestination is the templated upload destination in the test metadata.
const PublishDestination = "{{.Repository}}/testproject/{{.Version}}/linux/amd64/testproject"

// FileContent is the body of the file the checksum fixtures describe.
const FileContent = `the quick fox jumped over the lazy brown dog`

// FileMd5 is the md5 sum of FileContent.
const FileMd5 = "356b5768c6964531f678781446840b76"

// FileSha1 is the sha1 sum of FileContent.
const FileSha1 = "041b2390cd9697ba6b9f57b532b0aa5ac183736b"

// FileSha256 is the sha256 sum of FileContent.
const FileSha256 = "e088f8b9456b8a91a48159497ac425a4c3cdcad3ad81cc3a269618209dee033b"

// AllChecksums returns every checksum of FileContent, in the order gomason
// generates them: md5, sha1, sha256.
func AllChecksums() (checksums []string) {
	checksums = []string{FileMd5, FileSha1, FileSha256}

	return checksums
}

// UserConfig is a ~/.gomason carrying static and shell-function credentials.
const UserConfig = `[user]
  email = nik.ogura@gmail.com
  username = nikogura
  usernamefunc = echo 'foo bar baz'
  password = changeit
  passwordfunc = echo 'seecret!'

[signing]
  program = gpg
`

// UserConfigWithBearerToken is a ~/.gomason carrying a static bearer token.
const UserConfigWithBearerToken = `[user]
  email = nik.ogura@gmail.com
  username = nikogura
  password = changeit
  bearertoken = static-test-token-12345

[signing]
  program = gpg
`

// UserConfigWithBearerTokenFunc is a ~/.gomason that sources its bearer token
// from a shell function rather than a static value.
const UserConfigWithBearerTokenFunc = `[user]
  email = nik.ogura@gmail.com
  bearertokenfunc = echo 'dynamic-token-from-func'

[signing]
  program = gpg
`

// MetadataJSON is a metadata.json equivalent to the Metadata object the
// gomason tests build by hand.  Keep the two in step.
const MetadataJSON = `{
	"version": "0.1.0",
	"package": "github.com/nikogura/testproject",
	"description": "Test Project for Gomason.",
	"building": {
		"prepcommands": [
      "echo \"GOPATH is: ${GOPATH}\""
		],
		"targets": [
			{
				"name": "linux/amd64",
				"flags": {
					"FOO": "bar"
				}
			}
		]
	},
	"signing": {
		"program": "gpg",
		"email": "gomason-tester@foo.com"
	},
	"publishing": {
		"targets": [
			{
				"src": "testproject_linux_amd64",
				"dst": "{{.Repository}}/testproject/{{.Version}}/linux/amd64/testproject",
				"sig": true,
				"checksums": true
			}
		]
	}
}`

// BuildFlagName is the environment variable name set by the test build target.
const BuildFlagName = "FOO"

// BuildFlagValue is the value of BuildFlagName set by the test build target.
const BuildFlagValue = "bar"
