package gomason

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/phayes/freeport"

	"github.com/nikogura/gomason/pkg/logging"
)

var TestTmpDir string
var servicePort int

func TestMain(m *testing.M) {
	setUp()

	code := m.Run()

	tearDown()

	os.Exit(code)
}

func setUp() {
	logging.Init(true)

	dir, err := ioutil.TempDir("", "gomason")
	if err != nil {
		log.Fatal("Error creating temp dir\n")
	}

	TestTmpDir = dir

	log.Printf("Setting up temporary work dir %s", TestTmpDir)

	freePort, err := freeport.GetFreePort()
	if err != nil {
		log.Printf("Error getting a free port: %s", err)
		os.Exit(1)
	}

	servicePort = freePort

	tr := TestRepo{}

	go tr.Run(servicePort)

}

func tearDown() {
	if _, err := os.Stat(TestTmpDir); !os.IsNotExist(err) {
		_ = os.Remove(TestTmpDir)
	}
}

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
				{
					Name: "darwin/amd64",
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
					Source:      "testproject_darwin_amd64",
					Destination: "{{.Repository}}/testproject/{{.Version}}/darwin/amd64/testproject",
					Signature:   true,
					Checksums:   true,
				},
				{
					Source:      "testproject_linux_amd64",
					Destination: "{{.Repository}}/testproject/{{.Version}}/linux/amd64/testproject",
					Signature:   true,
					Checksums:   true,
				},
			},
			TargetsMap: map[string]PublishTarget{
				"testproject_darwin_amd64": {
					Source:      "testproject_darwin_amd64",
					Destination: "{{.Repository}}/testproject/{{.Version}}/darwin/amd64/testproject",
					Signature:   true,
					Checksums:   true,
				},
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

func TestGetUserConfig(t *testing.T) {
	homeDir := fmt.Sprintf("%s/user", TestTmpDir)

	err := os.Mkdir(homeDir, 0755)
	if err != nil {
		t.Errorf("Failed creating %s", homeDir)
	}

	userFile := fmt.Sprintf("%s/.gomason", homeDir)

	err = ioutil.WriteFile(userFile, []byte(testUserConfig()), 0644)
	if err != nil {
		t.Errorf("error writing %s: %s", userFile, err)
	}

	actual, err := GetUserConfig(homeDir)
	if err != nil {
		t.Errorf("Error getting user config: %s", err)
	}

	expected := UserConfig{
		User: UserInfo{
			Email:        "nik.ogura@gmail.com",
			Username:     "nikogura",
			UsernameFunc: "echo 'foo bar baz'",
			Password:     "changeit",
			PasswordFunc: "echo 'seecret!'",
		},
		Signing: UserSignInfo{
			Program: "gpg",
		},
	}

	assert.Equal(t, expected, actual, "loaded config meets expectations")
}
