package gomason

import (
	"fmt"
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
		_, statErr := os.Stat(filepath.Join(TestTmpDir, dir))
		if os.IsNotExist(statErr) {
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

	err = lang.Checkout(gopath, testMetadataObj(), "")
	if err != nil {
		t.Errorf("Failed to checkout module: %s", err)
	}

	metaPath := filepath.Join(gopath, "src", testModuleName(), MetadataFilename)
	_, statErr := os.Stat(metaPath)
	if os.IsNotExist(statErr) {
		t.Errorf("Failed to checkout module")
	}
}

func TestCheckoutBranch(t *testing.T) {
	// making a separate temp dir here cos it steps on the other tests
	dir := t.TempDir()

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
	_, statErr := os.Stat(testFilePath)
	if os.IsNotExist(statErr) {
		t.Errorf("Failed to checkout branch")
	}

}

func TestPrep(t *testing.T) {
	lang, _ := GetByName(LanguageGolang)
	gopath, err := lang.CreateWorkDir(TestTmpDir)
	if err != nil {
		t.Errorf("Error creating GOPATH in %s: %s", TestTmpDir, err)
	}

	err = lang.Checkout(gopath, testMetadataObj(), "")
	if err != nil {
		t.Errorf("Failed to checkout module: %s", err)
	}

	metaPath := filepath.Join(gopath, "src", testModuleName(), MetadataFilename)
	_, statErr := os.Stat(metaPath)
	if os.IsNotExist(statErr) {
		t.Errorf("Failed to checkout module")
	}

	err = lang.Prep(gopath, testMetadataObj(), false)
	if err != nil {
		t.Errorf("error running prep steps: %s", err)
	}
}

func TestBuildGoxInstall(t *testing.T) {
	lang, _ := GetByName(LanguageGolang)

	gopath, err := lang.CreateWorkDir(TestTmpDir)
	if err != nil {
		t.Errorf("Error creating GOPATH in %s: %s", TestTmpDir, err)
	}

	err = GoxInstall(gopath)
	if err != nil {
		t.Errorf("Error installing Gox: %s\n", err)
	}

	_, statErr := os.Stat(filepath.Join(gopath, "bin/gox"))
	if os.IsNotExist(statErr) {
		t.Errorf("Gox failed to install.")
	}
}

// setupBuildTest sets up a workspace for build tests: creates workdir, checks out code, and runs prep.
func setupBuildTest(t *testing.T, lang Language) (gopath string) {
	t.Helper()

	var err error

	gopath, err = lang.CreateWorkDir(TestTmpDir)
	if err != nil {
		t.Fatalf("Error creating GOPATH in %s: %s\n", TestTmpDir, err)
	}

	err = lang.Checkout(gopath, testMetadataObj(), "")
	if err != nil {
		t.Fatalf("Failed to checkout module: %s", err)
	}

	metaPath := filepath.Join(gopath, "src", testModuleName(), MetadataFilename)
	_, statErr := os.Stat(metaPath)
	if os.IsNotExist(statErr) {
		t.Fatalf("Failed to checkout module")
	}

	err = lang.Prep(gopath, testMetadataObj(), false)
	if err != nil {
		t.Fatalf("error running prep steps: %s", err)
	}

	return gopath
}

// verifyArtifactsPresent checks that the listed artifacts exist in workdir.
func verifyArtifactsPresent(t *testing.T, workdir string, artifacts []string) {
	t.Helper()

	for _, artifact := range artifacts {
		binary := fmt.Sprintf("%s/%s", workdir, artifact)

		_, binaryStatErr := os.Stat(binary)
		if os.IsNotExist(binaryStatErr) {
			t.Errorf("Gox failed to build binary: %s.\n", binary)
		}
	}
}

// verifyArtifactsMissing checks that the listed artifacts do NOT exist in workdir.
func verifyArtifactsMissing(t *testing.T, workdir string, artifacts []string) {
	t.Helper()

	for _, artifact := range artifacts {
		binary := fmt.Sprintf("%s/%s", workdir, artifact)

		_, binaryStatErr := os.Stat(binary)
		if os.IsNotExist(binaryStatErr) {
			fmt.Printf("Binary not found - as intended.\n")
		} else {
			t.Errorf("Gox built binary: %s when it shouldn't have.\n", binary)
		}
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
			[]string{},
			[]string{
				"testproject_linux_amd64",
			},
		},
		{
			"all-targets",
			LanguageGolang,
			"",
			[]string{
				"testproject_linux_amd64",
			},
			[]string{},
		},
	}

	for _, tc := range inputs {
		t.Run(tc.name, func(t *testing.T) {
			lang, err := GetByName(tc.lang)
			if err != nil {
				t.Error(err)
			}

			gopath := setupBuildTest(t, lang)
			gomodule := testMetadataObj().Package

			err = lang.Build(gopath, testMetadataObj(), tc.skipTargets, false)
			if err != nil {
				t.Errorf("Error building: %s", err)
			}

			workdir := filepath.Join(gopath, "src", gomodule)
			verifyArtifactsPresent(t, workdir, tc.artifactsPresent)
			verifyArtifactsMissing(t, workdir, tc.artifactsMissing)
		})
	}
}

func TestTest(t *testing.T) {
	lang, _ := GetByName(LanguageGolang)

	gopath, err := lang.CreateWorkDir(TestTmpDir)
	if err != nil {
		t.Errorf("Error creating GOPATH in %s: %s", TestTmpDir, err)
	}

	err = lang.Checkout(gopath, testMetadataObj(), "")
	if err != nil {
		t.Errorf("Failed to checkout module: %s", err)
	}

	metaPath := filepath.Join(gopath, "src", testModuleName(), MetadataFilename)
	_, statErr := os.Stat(metaPath)
	if os.IsNotExist(statErr) {
		t.Errorf("Failed to checkout module")
	}

	err = lang.Prep(gopath, testMetadataObj(), false)
	if err != nil {
		t.Errorf("error running prep steps: %s", err)
	}

	err = lang.Test(gopath, testMetadataObj().Package, "10m", false)
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
		t.Fatalf("Failed to check if gpg is installed:%s", err)
	}

	lang, _ := GetByName(LanguageGolang)

	// create workspace
	gopath, err := lang.CreateWorkDir(TestTmpDir)
	if err != nil {
		t.Fatalf("Error creating GOPATH in %s: %s\n", TestTmpDir, err)
	}

	err = lang.Checkout(gopath, testMetadataObj(), "")
	if err != nil {
		t.Fatalf("Failed to checkout module: %s", err)
	}

	meta := testMetadataObj()

	meta.Repository = fmt.Sprintf("http://localhost:%d/repo/tool", servicePort)

	// build artifacts
	err = lang.Build(gopath, meta, "", false)
	if err != nil {
		t.Fatalf("Error building: %s", err)
	}

	// set up GPG home directory (GPG 2.x requires --homedir to isolate all state)
	gpgHome := filepath.Join(TestTmpDir, "gnupg")
	err = os.MkdirAll(gpgHome, 0700)
	if err != nil {
		t.Fatalf("Error creating GPG home directory: %s", err)
	}

	meta.Options = make(map[string]interface{})
	meta.Options["homedir"] = gpgHome

	// write gpg batch file (explicit RSA for GPG 2.x compatibility)
	defaultKeyText := `%echo Generating a default key
%no-protection
%transient-key
Key-Type: RSA
Key-Length: 2048
Subkey-Type: RSA
Subkey-Length: 2048
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
		t.Fatalf("Error writing test key generation file: %s", err)
	}

	// generate a test key using --homedir for GPG 2.x compatibility
	cmd := exec.Command(shellCmd, "--homedir", gpgHome, "--batch", "--generate-key", keyFile)
	err = cmd.Run()
	if err != nil {
		t.Fatalf("****** Error creating test key: %s *****", err)
	}

	// sign binaries
	parts := strings.Split(meta.Package, "/")
	binaryPrefix := parts[len(parts)-1]

	for _, target := range meta.BuildInfo.Targets {
		archparts := strings.Split(target.Name, "/")

		osname := archparts[0]
		archname := archparts[1]

		workdir := filepath.Join(gopath, "src", meta.Package)
		binary := fmt.Sprintf("%s/%s_%s_%s", workdir, binaryPrefix, osname, archname)

		_, binaryStatErr := os.Stat(binary)
		if os.IsNotExist(binaryStatErr) {
			t.Errorf("Gox failed to build binary: %s\n", binary)
		}

		err = g.SignBinary(meta, binary)
		if err != nil {
			err = errors.Wrap(err, "failed to sign binary")
			t.Errorf("Failed to sign binary %s: %s", binary, err)
		}

		// verify binaries
		var ok bool
		ok, err = VerifyBinary(binary, meta)
		if err != nil {
			t.Errorf("Error verifying signature: %s", err)
		}

		if !ok {
			t.Errorf("Failed to verify signature on %s", binary)
		}

	}

	var cwd string
	cwd, err = os.Getwd()
	if err != nil {
		t.Errorf("Failed to get current working directory: %s", err)
	}

	err = g.HandleArtifacts(meta, gopath, cwd, false, true, true, "", false)
	if err != nil {
		t.Errorf("post-build processing failed: %s", err)
	}

	err = g.HandleExtras(meta, gopath, cwd, false, true, true, false)
	if err != nil {
		t.Errorf("Extra artifact processing failed: %s", err)
	}
}
