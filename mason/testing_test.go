package mason

import (
	"fmt"
	"log"
	"os"
	"testing"
)

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

	err = Checkout(gopath, testMetadataObj(), "master", true)
	if err != nil {
		log.Printf("Failed to checkout module: %s", err)
		t.Fail()
	}

	if _, err := os.Stat(fmt.Sprintf("%s/src/%s/metadata.json", gopath, testModuleName())); os.IsNotExist(err) {
		log.Printf("Failed to checkout module")
		t.Fail()
	}

	err = GovendorSync(gopath, testMetadataObj(), true)
	if err != nil {
		log.Printf("Error runnig govendor sync: %s", err)
		t.Fail()
	}
}
