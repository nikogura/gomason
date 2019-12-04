package languages

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/nikogura/gomason/pkg/gomason"
	"github.com/phayes/freeport"
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

	tr := gomason.TestRepo{}

	go tr.Run(servicePort)

}

func tearDown() {
	if _, err := os.Stat(TestTmpDir); !os.IsNotExist(err) {
		os.Remove(TestTmpDir)
	}
}
