package mason

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var tmpDir string

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

	tmpDir = dir

	log.Printf("Setting up temporary work dir %s", tmpDir)

}

func tearDown() {
	if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
		os.Remove(tmpDir)
	}
}

func TestWholeShebang(t *testing.T) {
	fileName := fmt.Sprintf("%s/%s", tmpDir, testMetadataFileName())

	err := ioutil.WriteFile(fileName, []byte(testMetaDataJson()), 0644)
	if err != nil {
		log.Printf("Error writing metadata file: %s", err)
		t.Fail()
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Printf("Error determining working directory: %s", wd)
		t.Fail()
	}

	os.Chdir(tmpDir)

	err = WholeShebang(tmpDir, "master", true, true, false, true)
	if err != nil {
		log.Printf("Error running whole shebang: %s")
		t.Fail()
	}

	os.Chdir(wd)
}
