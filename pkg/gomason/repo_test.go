package gomason

import (
	"fmt"
	"io/ioutil"
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

	// making a separate temp dir here cos it steps on the other tests
	dir, err := ioutil.TempDir("", "gomason")
	if err != nil {
		log.Fatal("Error creating temp dir\n")
	}

	gopath, err := CreateGoPath(dir)
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

	_ = os.Remove(dir)
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
