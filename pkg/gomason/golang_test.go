package gomason

import (
	"fmt"
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
		t.Errorf("Error creating gopath in %q: %s", TestTmpDir, err)
	}

	dirs := []string{"go", "go/src", "go/pkg", "go/bin"}

	for _, dir := range dirs {
		if _, err := os.Stat(filepath.Join(TestTmpDir, dir)); os.IsNotExist(err) {
			t.Errorf("GoPath not created.")
		}
	}
}

func TestCheckoutDefault(t *testing.T) {
	lang, _ := GetByName(LanguageGolang)
	gopath, err := lang.CreateWorkDir(TestTmpDir)
	if err != nil {
		t.Errorf("Error creating GOPATH in %s: %s", TestTmpDir, err)
	}

	log.Printf("Checking out Master Branch\n")
	err = lang.Checkout(gopath, testMetadataObj(), "")
	if err != nil {
		t.Errorf("Failed to checkout module: %s", err)
	}

	metaPath := filepath.Join(gopath, "src", testModuleName(), METADATA_FILENAME)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Errorf("Failed to checkout module")
	}
}

func TestCheckoutBranch(t *testing.T) {
	log.Printf("Checking out Test Branch\n")

	// making a separate temp dir here cos it steps on the other tests
	dir, err := os.MkdirTemp("", "gomason")
	if err != nil {
		t.Errorf("Error creating temp dir\n")
	}
	defer os.RemoveAll(dir)

	lang, _ := GetByName(LanguageGolang)
	gopath, err := lang.CreateWorkDir(dir)
	if err != nil {
		t.Errorf("Error creating GOPATH in %s: %s", dir, err)
	}

	err = lang.Checkout(gopath, testMetadataObj(), "testbranch")
	if err != nil {
		t.Errorf("Failed to checkout module: %s", err)
	}

	testFilePath := filepath.Join(gopath, "src", testModuleName(), "test_file")
	if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
		t.Errorf("Failed to checkout branch")
	}
}

func TestPrep(t *testing.T) {
	log.Printf("Checking out Master Branch\n")
	lang, _ := GetByName(LanguageGolang)
	gopath, err := lang.CreateWorkDir(TestTmpDir)
	if err != nil {
		t.Errorf("Error creating GOPATH in %s: %s", TestTmpDir, err)
	}

	err = lang.Checkout(gopath, testMetadataObj(), "")
	if err != nil {
		t.Errorf("Failed to checkout module: %s", err)
	}

	metaPath := filepath.Join(gopath, "src", testModuleName(), METADATA_FILENAME)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Errorf("Failed to checkout module")
	}

	err = lang.Prep(gopath, testMetadataObj())
	if err != nil {
		t.Errorf("error running prep steps: %s", err)
	}
}

func TestBuildGoxInstall(t *testing.T) {
	lang, _ := GetByName(LanguageGolang)

	log.Printf("Installing Gox\n")
	gopath, err := lang.CreateWorkDir(TestTmpDir)
	if err != nil {
		t.Errorf("Error creating GOPATH in %s: %s", TestTmpDir, err)
	}

	err = GoxInstall(gopath)
	if err != nil {
		t.Errorf("Error installing Gox: %s\n", err)
	}

	if _, err := os.Stat(filepath.Join(gopath, "bin/gox")); os.IsNotExist(err) {
		t.Errorf("Gox failed to install.")
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
				t.Errorf(err.Error())
			}

			log.Printf("Running Build\n")
			gopath, err := lang.CreateWorkDir(TestTmpDir)
			if err != nil {
				t.Errorf("Error creating GOPATH in %s: %s\n", TestTmpDir, err)
			}

			gomodule := testMetadataObj().Package

			log.Printf("Checking out Master Branch\n")

			err = lang.Checkout(gopath, testMetadataObj(), "")
			if err != nil {
				t.Errorf("Failed to checkout module: %s", err)
			}

			metaPath := filepath.Join(gopath, "src", testModuleName(), METADATA_FILENAME)
			if _, err := os.Stat(metaPath); os.IsNotExist(err) {
				t.Errorf("Failed to checkout module")
			}

			err = lang.Prep(gopath, testMetadataObj())
			if err != nil {
				t.Errorf("error running prep steps: %s", err)
			}

			err = lang.Build(gopath, testMetadataObj(), tc.skipTargets)
			if err != nil {
				t.Errorf("Error building: %s", err)
			}

			for _, artifact := range tc.artifactsPresent {
				workdir := filepath.Join(gopath, "src", gomodule)
				binary := fmt.Sprintf("%s/%s", workdir, artifact)

				log.Printf("Looking for binary present: %s\n", binary)

				if _, err := os.Stat(binary); os.IsNotExist(err) {
					t.Errorf("Gox failed to build binary: %s.\n", binary)
				} else {
					log.Printf("Binary found.\n")
				}
			}

			for _, artifact := range tc.artifactsMissing {
				workdir := filepath.Join(gopath, "src", gomodule)
				binary := fmt.Sprintf("%s/%s", workdir, artifact)

				log.Printf("Looking for binary not present: %s\n", binary)

				if _, err := os.Stat(binary); os.IsNotExist(err) {
					log.Printf("Binary not found - as intended.\n")
				} else {
					t.Errorf("Gox built binary: %s when it shouldn't have.\n", binary)
				}
			}
		})
	}
}

func TestTest(t *testing.T) {
	lang, _ := GetByName(LanguageGolang)

	log.Printf("Checking out Master Branch\n")
	gopath, err := lang.CreateWorkDir(TestTmpDir)
	if err != nil {
		t.Errorf("Error creating GOPATH in %s: %s", TestTmpDir, err)
	}

	err = lang.Checkout(gopath, testMetadataObj(), "")
	if err != nil {
		t.Errorf("Failed to checkout module: %s", err)
	}

	metaPath := filepath.Join(gopath, "src", testModuleName(), METADATA_FILENAME)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Errorf("Failed to checkout module")
	}

	err = lang.Prep(gopath, testMetadataObj())
	if err != nil {
		t.Errorf("error running prep steps: %s", err)
	}

	err = lang.Test(gopath, testMetadataObj().Package, "10m")
	if err != nil {
		t.Errorf("error running go test: %s", err)
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
		t.Errorf("Failed to check if gpg is installed:%s", err)
	}

	lang, _ := GetByName(LanguageGolang)

	// create workspace
	gopath, err := lang.CreateWorkDir(TestTmpDir)
	if err != nil {
		t.Errorf("Error creating GOPATH in %s: %s\n", TestTmpDir, err)
	}

	meta := testMetadataObj()

	meta.Repository = fmt.Sprintf("http://localhost:%d/repo/tool", servicePort)

	// build artifacts
	log.Printf("Running Build\n")
	err = lang.Build(gopath, meta, "")
	if err != nil {
		t.Errorf("Error building: %s", err)
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
	err = os.WriteFile(keyFile, []byte(defaultKeyText), 0644)
	if err != nil {
		t.Errorf("Error writing test key generation file: %s", err)
	}

	log.Printf("Keyring file: %s\n", keyring)
	log.Printf("Trustdb file: %s\n", trustdb)
	log.Printf("Test key generation file: %s\n", keyFile)

	// generate a test key
	cmd := exec.Command(shellCmd, "--trustdb", trustdb, "--no-default-keyring", "--keyring", keyring, "--batch", "--generate-key", keyFile)
	err = cmd.Run()
	if err != nil {
		t.Errorf("****** Error creating test key: %s *****", err)
	}

	log.Printf("Done creating keyring and test keys\n")

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
			t.Errorf("Gox failed to build binary: %s\n", binary)
		}

		err = g.SignBinary(meta, binary)
		if err != nil {
			err = errors.Wrap(err, "failed to sign binary")
			t.Errorf("Failed to sign binary %s: %s", binary, err)
		}

		// verify binaries
		ok, err := VerifyBinary(binary, meta)
		if err != nil {
			t.Errorf("Error verifying signature: %s", err)
		}

		if !ok {
			t.Errorf("Failed to verify signature on %s", binary)
		}

	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("Failed to get current working directory: %s", err)
	}

	fmt.Printf("Publishing\n")

	err = g.HandleArtifacts(meta, gopath, cwd, false, true, true, "")
	if err != nil {
		t.Errorf("post-build processing failed: %s", err)
	}

	err = g.HandleExtras(meta, gopath, cwd, false, true)
	if err != nil {
		t.Errorf("Extra artifact processing failed: %s", err)
	}
}
