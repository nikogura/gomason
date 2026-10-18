// Copyright © 2026 Nik Ogura <nik.ogura@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/nikogura/gomason/v2/pkg/gomason"
	"github.com/nikogura/gomason/v2/pkg/gomason/testfixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// prebuiltBOM is the bare CycloneDX BOM a fake syft emits in these tests.
const prebuiltBOM = `{"bomFormat": "CycloneDX", "specVersion": "1.6", "version": 1, "components": []}`

// shimPrebuiltSyft puts a syft shim on PATH that emits a bare CycloneDX BOM.
func shimPrebuiltSyft(t *testing.T) {
	t.Helper()

	dir := t.TempDir()

	script := filepath.Join(dir, "syft")
	contents := "#!/bin/sh\ncat <<'BOM'\n" + prebuiltBOM + "\nBOM\n"

	err := os.WriteFile(script, []byte(contents), 0755)
	require.NoError(t, err, "failed writing syft shim")

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// prebuiltGPGHome creates an isolated GPG home with a test signing key.
func prebuiltGPGHome(t *testing.T) (gpgHome string) {
	t.Helper()

	gpgHome = filepath.Join(t.TempDir(), "gnupg")

	err := os.MkdirAll(gpgHome, 0700)
	require.NoError(t, err, "failed creating GPG home directory")

	keyText := `%echo Generating a default key
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

	keyFile := filepath.Join(t.TempDir(), "testkey")
	err = os.WriteFile(keyFile, []byte(keyText), 0644)
	require.NoError(t, err, "failed writing key generation file")

	cmd := exec.Command("gpg", "--homedir", gpgHome, "--batch", "--generate-key", keyFile)
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	require.NoError(t, err, "failed generating test GPG key")

	return gpgHome
}

// prebuiltMetadata returns metadata matching the fixtures, with attestations enabled
// and publishing pointed at the given server.
func prebuiltMetadata(serverURL string) (meta gomason.Metadata) {
	target := gomason.PublishTarget{
		Source:       testfixtures.PublishSource,
		Destination:  testfixtures.PublishDestination,
		Signature:    true,
		Checksums:    true,
		Attestations: true,
	}

	meta = gomason.Metadata{
		Package:     testfixtures.ModuleName,
		Version:     testfixtures.Version,
		Description: testfixtures.Description,
		Repository:  fmt.Sprintf("%s/repo", serverURL),
		BuildInfo: gomason.BuildInfo{
			Attestations: true,
			Targets: []gomason.BuildTarget{
				{Name: testfixtures.BuildTargetName},
			},
		},
		SignInfo: gomason.SignInfo{
			Program: testfixtures.SigningProgram,
			Email:   testfixtures.SigningEmail,
		},
		PublishInfo: gomason.PublishInfo{
			Targets:    []gomason.PublishTarget{target},
			TargetsMap: map[string]gomason.PublishTarget{testfixtures.PublishSource: target},
		},
	}

	return meta
}

func TestPublishPrebuiltAttestations(t *testing.T) {
	shimPrebuiltSyft(t)

	var paths []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(http.StatusCreated)
	}))

	defer server.Close()

	// The prebuilt path works out of the cwd: sign and publish artifacts that were
	// built elsewhere.
	t.Chdir(t.TempDir())

	err := os.WriteFile(testfixtures.PublishSource, []byte("prebuilt bytes"), 0755)
	require.NoError(t, err, "failed writing prebuilt artifact")

	meta := prebuiltMetadata(server.URL)
	meta.Options = map[string]interface{}{"homedir": prebuiltGPGHome(t)}

	gm := &gomason.Gomason{}

	err = publishPrebuilt(gm, meta)
	require.NoError(t, err, "prebuilt publish with attestations must succeed")

	// Every file a signed, attested artifact carries must reach the repository —
	// including the attestations, which the prebuilt path must generate itself.
	expected := []string{
		"/repo/testproject/0.1.0/linux/amd64/testproject",
		"/repo/testproject/0.1.0/linux/amd64/testproject.asc",
		"/repo/testproject/0.1.0/linux/amd64/testproject.att.json",
		"/repo/testproject/0.1.0/linux/amd64/testproject.att.json.asc",
		"/repo/testproject/0.1.0/linux/amd64/testproject.sbom.json",
		"/repo/testproject/0.1.0/linux/amd64/testproject.sbom.json.asc",
	}

	for _, want := range expected {
		assert.Contains(t, paths, want, "prebuilt publish must upload %s", want)
	}
}

func TestPublishPrebuiltAttestationsUnmatchedTargetRefuses(t *testing.T) {
	shimPrebuiltSyft(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	defer server.Close()

	t.Chdir(t.TempDir())

	// An artifact whose name maps to no build target (no os/arch suffix to resolve):
	// provenance for it would have to carry a blank target field, which is a quiet
	// lie. The prebuilt path must refuse rather than attest a binary it cannot
	// describe.
	const unmatchedSource = "mystery_binary"

	err := os.WriteFile(unmatchedSource, []byte("prebuilt bytes"), 0755)
	require.NoError(t, err, "failed writing prebuilt artifact")

	meta := prebuiltMetadata(server.URL)
	meta.Options = map[string]interface{}{"homedir": prebuiltGPGHome(t)}

	target := meta.PublishInfo.Targets[0]
	target.Source = unmatchedSource
	meta.PublishInfo.Targets = []gomason.PublishTarget{target}
	meta.PublishInfo.TargetsMap = map[string]gomason.PublishTarget{unmatchedSource: target}

	gm := &gomason.Gomason{}

	err = publishPrebuilt(gm, meta)
	require.Error(t, err, "prebuilt publish must refuse an artifact with no matching build target")
	assert.Contains(t, err.Error(), "build target", "the error must say what could not be resolved")
}
