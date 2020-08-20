package gomason

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

func TestCreateGoPath(t *testing.T) {
	lang, _ := GetByName(LanguageGolang)
	_, err := lang.CreateWorkDir(TestTmpDir)
	if err != nil {
		log.Printf("Error creating gopath in %q: %s", TestTmpDir, err)
		t.FailNow()
	}

	dirs := []string{"go", "go/src", "go/pkg", "go/bin"}

	for _, dir := range dirs {
		if _, err := os.Stat(filepath.Join(TestTmpDir, dir)); os.IsNotExist(err) {
			t.Fail()
		}
	}
}

func TestCheckoutDefault(t *testing.T) {
	lang, _ := GetByName(LanguageGolang)
	gopath, err := lang.CreateWorkDir(TestTmpDir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s", TestTmpDir, err)
		t.FailNow()
	}

	log.Printf("Checking out Master Branch")
	err = lang.Checkout(gopath, testMetadataObj(), "master")
	if err != nil {
		log.Printf("Failed to checkout module: %s", err)
		t.FailNow()
	}

	metaPath := filepath.Join(gopath, "src", testModuleName(), METADATA_FILENAME)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		log.Printf("Failed to checkout module")
		t.FailNow()
	}
}

func TestCheckoutBranch(t *testing.T) {
	log.Printf("Checking out Test Branch")

	// making a separate temp dir here cos it steps on the other tests
	dir, err := ioutil.TempDir("", "gomason")
	if err != nil {
		log.Fatal("Error creating temp dir\n")
	}
	defer os.RemoveAll(dir)

	lang, _ := GetByName(LanguageGolang)
	gopath, err := lang.CreateWorkDir(dir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s", dir, err)
		t.FailNow()
	}

	err = lang.Checkout(gopath, testMetadataObj(), "testbranch")
	if err != nil {
		log.Printf("Failed to checkout module: %s", err)
		t.FailNow()
	}

	testFilePath := filepath.Join(gopath, "src", testModuleName(), "test_file")
	if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
		log.Printf("Failed to checkout branch")
		t.FailNow()
	}
}

func TestPrep(t *testing.T) {
	log.Printf("Checking out Master Branch")
	lang, _ := GetByName(LanguageGolang)
	gopath, err := lang.CreateWorkDir(TestTmpDir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s", TestTmpDir, err)
		t.FailNow()
	}

	err = lang.Checkout(gopath, testMetadataObj(), "master")
	if err != nil {
		log.Printf("Failed to checkout module: %s", err)
		t.FailNow()
	}

	metaPath := filepath.Join(gopath, "src", testModuleName(), METADATA_FILENAME)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		log.Printf("Failed to checkout module")
		t.FailNow()
	}

	err = lang.Prep(gopath, testMetadataObj())
	if err != nil {
		log.Printf("error running prep steps: %s", err)
		t.FailNow()
	}
}

func TestBuildGoxInstall(t *testing.T) {
	lang, _ := GetByName(LanguageGolang)

	log.Printf("Installing Gox\n")
	gopath, err := lang.CreateWorkDir(TestTmpDir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s", TestTmpDir, err)
		t.FailNow()
	}

	err = GoxInstall(gopath)
	if err != nil {
		log.Printf("Error installing Gox: %s\n", err)
		t.FailNow()
	}

	if _, err := os.Stat(filepath.Join(gopath, "bin/gox")); os.IsNotExist(err) {
		log.Printf("Gox failed to install.")
		t.FailNow()
	}
}

func TestBuild(t *testing.T) {
	inputs := []struct {
		name             string
		lang             string
		skipTargets      string
		artifactsPresent []string
		artifactsMissing []string
	}{
		{
			"skip-linux",
			LanguageGolang,
			"linux/amd64",
			[]string{
				"testproject_darwin_amd64",
			},
			[]string{
				"testproject_linux_amd64",
			},
		},
		{
			"all-targets",
			LanguageGolang,
			"",
			[]string{
				"testproject_darwin_amd64",
				"testproject_linux_amd64",
			},
			[]string{},
		},
	}

	for _, tc := range inputs {
		t.Run(tc.name, func(t *testing.T) {
			lang, err := GetByName(tc.lang)
			if err != nil {
				t.Fatalf(err.Error())
			}

			log.Printf("Running Build\n")
			gopath, err := lang.CreateWorkDir(TestTmpDir)
			if err != nil {
				log.Printf("Error creating GOPATH in %s: %s\n", TestTmpDir, err)
				t.FailNow()
			}

			gomodule := testMetadataObj().Package

			log.Printf("Checking out Master Branch")

			err = lang.Checkout(gopath, testMetadataObj(), "master")
			if err != nil {
				log.Printf("Failed to checkout module: %s", err)
				t.FailNow()
			}

			metaPath := filepath.Join(gopath, "src", testModuleName(), METADATA_FILENAME)
			if _, err := os.Stat(metaPath); os.IsNotExist(err) {
				log.Printf("Failed to checkout module")
				t.FailNow()
			}

			err = lang.Prep(gopath, testMetadataObj())
			if err != nil {
				log.Printf("error running prep steps: %s", err)
				t.FailNow()
			}

			err = lang.Build(gopath, testMetadataObj(), tc.skipTargets)
			if err != nil {
				log.Printf("Error building: %s", err)
				t.FailNow()
			}

			for _, artifact := range tc.artifactsPresent {
				workdir := filepath.Join(gopath, "src", gomodule)
				binary := fmt.Sprintf("%s/%s", workdir, artifact)

				log.Printf("Looking for binary present: %s", binary)

				if _, err := os.Stat(binary); os.IsNotExist(err) {
					log.Printf("Gox failed to build binary: %s.\n", binary)
					t.FailNow()
				} else {
					log.Printf("Binary found.")
				}
			}

			for _, artifact := range tc.artifactsMissing {
				workdir := filepath.Join(gopath, "src", gomodule)
				binary := fmt.Sprintf("%s/%s", workdir, artifact)

				log.Printf("Looking for binary not present: %s", binary)

				if _, err := os.Stat(binary); os.IsNotExist(err) {
					log.Printf("Binary not found - as intended.")
				} else {
					log.Printf("Gox built binary: %s when it shouldn't have.\n", binary)
					t.FailNow()
				}
			}
		})
	}
}

func TestTest(t *testing.T) {
	lang, _ := GetByName(LanguageGolang)

	log.Printf("Checking out Master Branch")
	gopath, err := lang.CreateWorkDir(TestTmpDir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s", TestTmpDir, err)
		t.FailNow()
	}

	err = lang.Checkout(gopath, testMetadataObj(), "master")
	if err != nil {
		log.Printf("Failed to checkout module: %s", err)
		t.FailNow()
	}

	metaPath := filepath.Join(gopath, "src", testModuleName(), METADATA_FILENAME)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		log.Printf("Failed to checkout module")
		t.FailNow()
	}

	err = lang.Prep(gopath, testMetadataObj())
	if err != nil {
		log.Printf("error running prep steps: %s", err)
		t.FailNow()
	}

	err = lang.Test(gopath, testMetadataObj().Package)
	if err != nil {
		log.Printf("error running go test: %s", err)
		t.FailNow()
	}
}

func TestSignVerifyBinary(t *testing.T) {
	g := Gomason{
		Config: UserConfig{
			User:    UserInfo{},
			Signing: UserSignInfo{},
		},
	}
	shellCmd, err := exec.LookPath("gpg")
	if err != nil {
		log.Printf("Failed to check if gpg is installed:%s", err)
		t.FailNow()
	}

	lang, _ := GetByName(LanguageGolang)

	// create workspace
	gopath, err := lang.CreateWorkDir(TestTmpDir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s\n", TestTmpDir, err)
		t.FailNow()
	}

	meta := testMetadataObj()

	meta.Repository = fmt.Sprintf("http://localhost:%d/repo/tool", servicePort)

	// build artifacts
	log.Printf("Running Build\n")
	err = lang.Build(gopath, meta, "")
	if err != nil {
		log.Printf("Error building: %s", err)
		t.FailNow()
	}

	// set up test keys
	keyring := filepath.Join(TestTmpDir, "keyring.gpg")
	trustdb := filepath.Join(TestTmpDir, "trustdb.gpg")

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
	keyFile := filepath.Join(TestTmpDir, "testkey")
	err = ioutil.WriteFile(keyFile, []byte(defaultKeyText), 0644)
	if err != nil {
		log.Printf("Error writing test key generation file: %s", err)
		t.FailNow()
	}

	log.Printf("Keyring file: %s", keyring)
	log.Printf("Trustdb file: %s", trustdb)
	log.Printf("Test key generation file: %s", keyFile)

	// generate a test key
	cmd := exec.Command(shellCmd, "--trustdb", trustdb, "--no-default-keyring", "--keyring", keyring, "--batch", "--generate-key", keyFile)
	err = cmd.Run()
	if err != nil {
		log.Printf("****** Error creating test key: %s *****", err)
		t.FailNow()
	}

	log.Printf("Done creating keyring and test keys")

	// sign binaries
	parts := strings.Split(meta.Package, "/")
	binaryPrefix := parts[len(parts)-1]

	for _, target := range meta.BuildInfo.Targets {
		archparts := strings.Split(target.Name, "/")

		osname := archparts[0]   // linux or darwin generally
		archname := archparts[1] // amd64 generally

		workdir := filepath.Join(gopath, "src", meta.Package)
		binary := fmt.Sprintf("%s/%s_%s_%s", workdir, binaryPrefix, osname, archname)

		if _, err := os.Stat(binary); os.IsNotExist(err) {
			fmt.Printf("Gox failed to build binary: %s\n", binary)
			log.Printf("Failed to find binary %s", binary)
			t.FailNow()
		}

		err = g.SignBinary(meta, binary)
		if err != nil {
			err = errors.Wrap(err, "failed to sign binary")
			log.Printf("Failed to sign binary %s: %s", binary, err)
			t.FailNow()
		}

		// verify binaries
		ok, err := VerifyBinary(binary, meta)
		if err != nil {
			log.Printf("Error verifying signature: %s", err)
			//t.Fail()
		}

		if !ok {
			log.Printf("Failed to verify signature on %s", binary)
			t.FailNow()
		}

	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %s", err)
	}

	fmt.Printf("Publishing\n")

	err = g.HandleArtifacts(meta, gopath, cwd, false, true, true, "")
	if err != nil {
		log.Fatalf("post-build processing failed: %s", err)
	}

	err = g.HandleExtras(meta, gopath, cwd, false, true)
	if err != nil {
		log.Fatalf("Extra artifact processing failed: %s", err)
	}
}
