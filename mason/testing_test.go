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

func TestCheckoutDefault(t *testing.T) {
	log.Printf("Checking out Master Branch")
	gopath, err := CreateGoPath(tmpDir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s", tmpDir, err)
		t.Fail()
	}

	err = Checkout(gopath, testModuleName(), "master", true)
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

	err = Checkout(gopath, testModuleName(), "test_branch", true)
	if err != nil {
		log.Printf("Failed to checkout module: %s", err)
		t.Fail()
	}

	if _, err := os.Stat(fmt.Sprintf("%s/src/%s/test_file", gopath, testModuleName())); os.IsNotExist(err) {
		log.Printf("Failed to checkout branch")
		t.Fail()
	}

}

func TestGovendorInstall(t *testing.T) {
	log.Printf("Installing Govendor")
	gopath, err := CreateGoPath(tmpDir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s", tmpDir, err)
		t.Fail()
	}

	err = GovendorInstall(gopath, true)
	if err != nil {
		log.Printf("Error installing Govendor: %s", err)
		t.Fail()
	}

	if _, err := os.Stat(fmt.Sprintf("%s/go/bin/govendor", tmpDir)); os.IsNotExist(err) {
		log.Printf("Govendor vailed to install.")
		t.Fail()
	}

}

func TestGovendorSync(t *testing.T) {
	log.Printf("Installing Govendor")
	gopath, err := CreateGoPath(tmpDir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s", tmpDir, err)
		t.Fail()
	}

	err = GovendorInstall(gopath, true)
	if err != nil {
		log.Printf("Error installing Govendor: %s", err)
		t.Fail()
	}

	if _, err := os.Stat(fmt.Sprintf("%s/go/bin/govendor", tmpDir)); os.IsNotExist(err) {
		log.Printf("Govendor vailed to install.")
		t.Fail()
	}

	log.Printf("Checking out Master Branch")

	err = Checkout(gopath, testModuleName(), "master", true)
	if err != nil {
		log.Printf("Failed to checkout module: %s", err)
		t.Fail()
	}

	if _, err := os.Stat(fmt.Sprintf("%s/src/%s/metadata.json", gopath, testModuleName())); os.IsNotExist(err) {
		log.Printf("Failed to checkout module")
		t.Fail()
	}

	err = GovendorSync(gopath, "github.com/nikogura/gomason", true)
	if err != nil {
		log.Printf("Error runnig govendor sync: %s", err)
		t.Fail()
	}
}
