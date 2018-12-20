package gomason

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestBuildGoxInstall(t *testing.T) {
	log.Printf("Installing Go\n")
	gopath, err := CreateGoPath(tmpDir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s", tmpDir, err)
		t.Fail()
	}

	err = GoxInstall(gopath, true)
	if err != nil {
		log.Printf("Error installing Gox: %s\n", err)
		t.Fail()
	}

	if _, err := os.Stat(fmt.Sprintf("%s/go/bin/gox", tmpDir)); os.IsNotExist(err) {
		log.Printf("Gox failed to install.")
		t.Fail()
	}
}

func TestBuild(t *testing.T) {
	log.Printf("Running Build\n")
	gopath, err := CreateGoPath(tmpDir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s\n", tmpDir, err)
		t.Fail()
	}

	gomodule := testMetadataObj().Package
	branch := "master"

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

	err = Prep(gopath, testMetadataObj(), true)
	if err != nil {
		log.Printf("error running prep steps: %s", err)
		t.Fail()
	}

	err = Build(gopath, testMetadataObj(), branch, true)
	if err != nil {
		log.Printf("Error building: %s", err)
		t.Fail()
	}

	parts := strings.Split(gomodule, "/")

	binaryPrefix := parts[len(parts)-1]

	osname := runtime.GOOS
	archname := runtime.GOARCH

	workdir := fmt.Sprintf("%s/src/%s", gopath, gomodule)
	binary := fmt.Sprintf("%s/%s_%s_%s", workdir, binaryPrefix, osname, archname)

	log.Printf("Looking for binary: %s", binary)

	if _, err := os.Stat(binary); os.IsNotExist(err) {
		log.Printf("Gox failed to build binary: %s.\n", binary)
		t.Fail()
	} else {
		log.Printf("Binary found.")
	}
}
