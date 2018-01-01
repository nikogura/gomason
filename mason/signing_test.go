package mason

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestSignBinary(t *testing.T) {
	shellCmd, err := exec.LookPath("gpg")
	if err != nil {
		log.Printf("Failed to check if gpg is installed:%s", err)
		t.Fail()
	}

	meta, err := ReadMetadata("metadata.json")

	// create workspace
	gopath, err := CreateGoPath(tmpDir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s\n", tmpDir, err)
		t.Fail()
	}

	gomodule := "github.com/nikogura/gomason"
	branch := "master"

	// build artifacts
	log.Printf("Running Build\n")
	err = Build(gopath, gomodule, branch, true)
	if err != nil {
		log.Printf("Error building: %s", err)
		t.Fail()
	}

	// set up test keys
	keyring := fmt.Sprintf("%s/keyring.gpg", tmpDir)
	trustdb := fmt.Sprintf("%s/trustdb.gpg", tmpDir)

	meta.Options = make(map[string]interface{})
	meta.Options["keyring"] = keyring
	meta.Options["trustdb"] = trustdb

	// write gpg batch file
	defaultKeyText := `%echo Generating a default key
%no-protection
%transient-key
Key-Type: default
Subkey-Type: default
Name-Real: Gomason Tester
Name-Comment: with no passphrase
Name-Email: gomason-tester@foo.com
Expire-Date: 0
%commit
%echo done
`
	keyFile := fmt.Sprintf("%s/testkey", tmpDir)
	err = ioutil.WriteFile(keyFile, []byte(defaultKeyText), 0644)
	if err != nil {
		log.Printf("Error writing test key generation file: %s", err)
		t.Fail()
	}

	log.Printf("Keyring file: %s", keyring)
	log.Printf("Trustdb file: %s", trustdb)
	log.Printf("Test key generation file: %s", keyFile)

	// generate a test key
	cmd := exec.Command(shellCmd, "--trustdb", trustdb, "--no-default-keyring", "--keyring", keyring, "--batch", "--generate-key", keyFile)
	err = cmd.Run()
	if err != nil {
		log.Printf("****** Error creating test key: %s *****", err)
		t.Fail()
	}

	// sign binaries
	parts := strings.Split(meta.Package, "/")
	binaryPrefix := parts[len(parts)-1]

	for _, arch := range meta.BuildInfo.Targets {
		archparts := strings.Split(arch, "/")

		osname := archparts[0]   // linux or darwin generally
		archname := archparts[1] // amd64 generally

		workdir := fmt.Sprintf("%s/src/%s", gopath, meta.Package)
		binary := fmt.Sprintf("%s/%s_%s_%s", workdir, binaryPrefix, osname, archname)

		if _, err := os.Stat(binary); os.IsNotExist(err) {
			fmt.Printf("Gox failed to build binary: %s\n", binary)
			log.Printf("Failed to find binary %s", binary)
			t.Fail()
		}

		// sign 'em if we're signing
		err = SignBinary(meta, binary, true)
		if err != nil {
			err = errors.Wrap(err, "failed to sign binary")
			log.Printf("Failed to sign binary %s", binary)
			t.Fail()
		}

		// verify binaries
		ok, err := VerifyBinary(binary, meta)
		if err != nil {
			log.Printf("Error verifying signature: %s", err)
			//t.Fail()
		}

		if !ok {
			log.Printf("Failed to verify signature on %s", binary)
			t.Fail()
		}

	}

}
