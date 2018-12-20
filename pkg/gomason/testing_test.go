package gomason

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func TestTest(t *testing.T) {
	log.Printf("Checking out Master Branch")
	gopath, err := CreateGoPath(tmpDir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s", tmpDir, err)
		t.Fail()
	}

	err = Checkout(gopath, testMetadataObj(), "master", true)
	if err != nil {
		log.Printf("Failed to checkout module: %s", err)
		t.Fail()
	}

	if _, err := os.Stat(fmt.Sprintf("%s/src/%s/metadata.json", gopath, testModuleName())); os.IsNotExist(err) {
		log.Printf("Failed to checkout module")
		t.Fail()
	}

	err = Prep(gopath, testMetadataObj(), true)
	if err != nil {
		log.Printf("error running prep steps: %s", err)
		t.Fail()
	}

	err = GoTest(gopath, testMetadataObj().Package, true)
	if err != nil {
		log.Printf("error running go test: %s", err)
		t.Fail()
	}
}
