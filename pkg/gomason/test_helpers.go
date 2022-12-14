package gomason

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// testMetadataObj returns a Metadata object suitable for testing
func testMetadataObj() (metadata Metadata) {
	metadata = Metadata{
		Package:     testModuleName(),
		Version:     "0.1.0",
		Description: "Test Project for Gomason.",
		BuildInfo: BuildInfo{
			PrepCommands: []string{
				"echo \"GOPATH is: ${GOPATH}\"",
			},
			Targets: []BuildTarget{
				{
					Name: "linux/amd64",
					Flags: map[string]string{
						"FOO": "bar",
					},
				},
			},
		},
		SignInfo: SignInfo{
			Program: "gpg",
			Email:   "gomason-tester@foo.com",
		},
		PublishInfo: PublishInfo{
			Targets: []PublishTarget{
				{
					Source:      "testproject_linux_amd64",
					Destination: "{{.Repository}}/testproject/{{.Version}}/linux/amd64/testproject",
					Signature:   true,
					Checksums:   true,
				},
			},
			TargetsMap: map[string]PublishTarget{
				"testproject_linux_amd64": {
					Source:      "testproject_linux_amd64",
					Destination: "{{.Repository}}/testproject/{{.Version}}/linux/amd64/testproject",
					Signature:   true,
					Checksums:   true,
				},
			},
		},
	}

	return metadata
}

// testModuleName returns the name of the test module
func testModuleName() string {
	return "github.com/nikogura/testproject"
}

func TestMetadata_GetLanguage(t *testing.T) {
	cases := []struct {
		name     string
		meta     Metadata
		expected string
	}{
		{
			"golang",
			Metadata{
				Language: "golang",
			},
			"golang",
		},
		{
			"python",
			Metadata{
				Language: "python",
			},
			"python",
		},
		{
			"default",
			Metadata{
				Language: "",
			},
			"golang",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := c.meta.GetLanguage()

			assert.Equal(t, c.expected, actual, "returned language meets expectations")
		})
	}
}

func testUserConfig() string {
	return `[user]
  email = nik.ogura@gmail.com
  username = nikogura
  usernamefunc = echo 'foo bar baz'
  password = changeit
  passwordfunc = echo 'seecret!'

[signing]
  program = gpg
`
}

func testMetaDataJson() string {
	return `{
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
}

func testMetadataFileName() string {
	return METADATA_FILENAME
}

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
