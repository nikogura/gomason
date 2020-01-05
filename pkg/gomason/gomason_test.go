package gomason

import (
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
			Targets: []BuildTarget{{Name: "linux/amd64"}, {Name: "darwin/amd64"}},
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
