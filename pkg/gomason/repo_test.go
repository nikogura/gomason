package gomason

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func TestCheckoutDefault(t *testing.T) {
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
}

func TestCheckoutBranch(t *testing.T) {
	log.Printf("Checking out Test Branch")
	gopath, err := CreateGoPath(tmpDir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s", tmpDir, err)
		t.Fail()
	}

	err = Checkout(gopath, testMetadataObj(), "testbranch", true)
	if err != nil {
		log.Printf("Failed to checkout module: %s", err)
		t.Fail()
	}

	if _, err := os.Stat(fmt.Sprintf("%s/src/%s/test_file", gopath, testModuleName())); os.IsNotExist(err) {
		log.Printf("Failed to checkout branch")
		t.Fail()
	}
}

func TestPrep(t *testing.T) {
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
}
