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

package gomason

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nikogura/gomason/v2/pkg/gomason/testfixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// minimalCycloneDXBOM is a bare CycloneDX BOM, as syft emits with -o cyclonedx-json.
const minimalCycloneDXBOM = `{
  "bomFormat": "CycloneDX",
  "specVersion": "1.6",
  "version": 1,
  "components": [
    {
      "type": "library",
      "name": "github.com/spf13/cobra",
      "version": "v1.8.1",
      "purl": "pkg:golang/github.com/spf13/cobra@v1.8.1"
    }
  ]
}`

// writeTempArtifact writes content to a temp file and returns its path plus the file's sha256.
func writeTempArtifact(t *testing.T, content string) (path string, sha256 string) {
	t.Helper()

	path = filepath.Join(t.TempDir(), "testproject_linux_amd64")

	err := os.WriteFile(path, []byte(content), 0755)
	require.NoError(t, err, "failed writing temp artifact")

	sha256, err = FileSha256(path)
	require.NoError(t, err, "failed checksumming temp artifact")

	return path, sha256
}

func TestGenerateProvenanceStatement(t *testing.T) {
	t.Parallel()

	meta := testMetadataObj()
	meta.BuildInfo.Attestations = true

	input := provenanceInput{
		meta:      meta,
		target:    "linux/amd64",
		subject:   "testproject",
		sha256:    testfixtures.FileSha256,
		sourceURI: "git+https://github.com/nikogura/testproject",
		commit:    strings.Repeat("a1b2c3d4", 5), // 40 hex chars: a sha1 commit
	}

	stmt, err := generateProvenanceStatement(input)
	require.NoError(t, err, "provenance generation must succeed")

	assert.Equal(t, "https://in-toto.io/Statement/v1", stmt.Type, "statement _type")
	assert.Equal(t, "https://slsa.dev/provenance/v1", stmt.PredicateType, "statement predicateType")
	require.Len(t, stmt.Subject, 1, "exactly one subject")
	assert.Equal(t, "testproject", stmt.Subject[0].Name, "subject name is the destination binary name")
	assert.Equal(t, testfixtures.FileSha256, stmt.Subject[0].Digest["sha256"], "subject digest is the binary checksum")

	// Round-trip the statement through JSON and assert the emitted predicate fields —
	// this also guards the json tags against silent drift.
	encoded, err := json.Marshal(stmt)
	require.NoError(t, err, "statement must marshal")

	decoded := struct {
		Predicate struct {
			BuildDefinition struct {
				BuildType          string `json:"buildType"`
				ExternalParameters struct {
					Package string `json:"package"`
					Version string `json:"version"`
					Target  string `json:"target"`
				} `json:"externalParameters"`
				ResolvedDependencies []struct {
					URI    string            `json:"uri"`
					Digest map[string]string `json:"digest"`
				} `json:"resolvedDependencies"`
			} `json:"buildDefinition"`
			RunDetails struct {
				Builder struct {
					ID string `json:"id"`
				} `json:"builder"`
				Metadata struct {
					InvocationID string `json:"invocationId"`
				} `json:"metadata"`
			} `json:"runDetails"`
		} `json:"predicate"`
	}{}

	err = json.Unmarshal(encoded, &decoded)
	require.NoError(t, err, "marshaled statement must be valid JSON with the expected shape")

	assert.Equal(t, "https://gomason.nikogura.com/build/v1", decoded.Predicate.BuildDefinition.BuildType, "buildType")
	assert.Equal(t, testfixtures.ModuleName, decoded.Predicate.BuildDefinition.ExternalParameters.Package, "package parameter")
	assert.Equal(t, testfixtures.Version, decoded.Predicate.BuildDefinition.ExternalParameters.Version, "version parameter")
	assert.Equal(t, "linux/amd64", decoded.Predicate.BuildDefinition.ExternalParameters.Target, "target parameter")

	require.Len(t, decoded.Predicate.BuildDefinition.ResolvedDependencies, 1, "one resolved dependency")
	assert.Equal(t, "git+https://github.com/nikogura/testproject", decoded.Predicate.BuildDefinition.ResolvedDependencies[0].URI, "source URI")
	assert.Equal(t, strings.Repeat("a1b2c3d4", 5), decoded.Predicate.BuildDefinition.ResolvedDependencies[0].Digest["sha1"], "source commit digest")

	assert.Equal(t, testfixtures.SigningEmail, decoded.Predicate.RunDetails.Builder.ID, "builder id is the signing identity")
	assert.NotEmpty(t, decoded.Predicate.RunDetails.Metadata.InvocationID, "invocation id is generated")
}

func TestGenerateProvenanceStatementNoCommit(t *testing.T) {
	t.Parallel()

	// A workdir without git info (local build, or detached export) still produces a
	// statement; the dependency carries the URI without a digest rather than failing.
	meta := testMetadataObj()

	input := provenanceInput{
		meta:      meta,
		target:    "linux/amd64",
		subject:   "testproject",
		sha256:    testfixtures.FileSha256,
		sourceURI: "git+https://github.com/nikogura/testproject",
		commit:    "",
	}

	stmt, err := generateProvenanceStatement(input)
	require.NoError(t, err, "provenance generation without a commit must succeed")

	encoded, err := json.Marshal(stmt)
	require.NoError(t, err, "statement must marshal")

	decoded := struct {
		Predicate struct {
			BuildDefinition struct {
				ResolvedDependencies []struct {
					URI    string            `json:"uri"`
					Digest map[string]string `json:"digest"`
				} `json:"resolvedDependencies"`
			} `json:"buildDefinition"`
		} `json:"predicate"`
	}{}

	err = json.Unmarshal(encoded, &decoded)
	require.NoError(t, err, "marshaled statement must be valid JSON")

	require.Len(t, decoded.Predicate.BuildDefinition.ResolvedDependencies, 1, "dependency list keeps the source URI")
	assert.Empty(t, decoded.Predicate.BuildDefinition.ResolvedDependencies[0].Digest, "no digest is claimed without a commit")
}

func TestGenerateSBOMStatement(t *testing.T) {
	t.Parallel()

	stmt := generateSBOMStatement("testproject", testfixtures.FileSha256, json.RawMessage(minimalCycloneDXBOM))

	assert.Equal(t, "https://in-toto.io/Statement/v1", stmt.Type, "statement _type")
	assert.Equal(t, "https://cyclonedx.org/bom", stmt.PredicateType, "statement predicateType")
	require.Len(t, stmt.Subject, 1, "exactly one subject")
	assert.Equal(t, "testproject", stmt.Subject[0].Name, "subject name is the destination binary name")
	assert.Equal(t, testfixtures.FileSha256, stmt.Subject[0].Digest["sha256"], "subject digest is the binary checksum")

	// The BOM must be carried verbatim in the predicate.
	bom := struct {
		BOMFormat   string `json:"bomFormat"`
		SpecVersion string `json:"specVersion"`
	}{}

	err := json.Unmarshal(stmt.Predicate, &bom)
	require.NoError(t, err, "predicate must carry the BOM verbatim")

	assert.Equal(t, "CycloneDX", bom.BOMFormat, "bom format")
	assert.Equal(t, "1.6", bom.SpecVersion, "bom spec version")
}

func TestAttestationSubjectName(t *testing.T) {
	t.Parallel()

	meta := testMetadataObj()
	meta.Repository = "http://example.com/repo"

	t.Run("from publish destination", func(t *testing.T) {
		t.Parallel()
		subject, err := attestationSubjectName(meta, testfixtures.PublishSource)
		require.NoError(t, err, "matching target must yield a subject name")

		assert.Equal(t, "testproject", subject, "subject is the final path segment of the destination")
	})

	t.Run("no matching publish target is an error", func(t *testing.T) {
		t.Parallel()

		_, err := attestationSubjectName(meta, "some_other_artifact")
		require.Error(t, err, "attestations without a publish target must not guess a subject name")
	})
}

func TestReadMetadataWithAttestations(t *testing.T) {
	t.Parallel()

	contents := strings.Replace(testfixtures.MetadataJSON, `"building": {`,
		`"building": {
		"attestations": true,`, 1)

	contents = strings.Replace(contents, `"sig": true,`,
		`"sig": true,
				"attestations": true,`, 1)

	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "metadata.json"), []byte(contents), 0644)
	require.NoError(t, err, "failed writing metadata fixture")

	meta, err := ReadMetadata(filepath.Join(dir, "metadata.json"))
	require.NoError(t, err, "metadata with attestations must parse")

	assert.True(t, meta.BuildInfo.Attestations, "building.attestations must round-trip")

	require.Len(t, meta.PublishInfo.Targets, 1, "one publish target")
	assert.True(t, meta.PublishInfo.Targets[0].Attestations, "target attestations must round-trip")
}

// writeFakeSyft drops a syft shim on PATH that emits the given bare CycloneDX BOM on
// stdout, standing in for the real tool so the generation path is tested deterministically.
func writeFakeSyft(t *testing.T, bom string) {
	t.Helper()

	dir := t.TempDir()

	script := filepath.Join(dir, "syft")
	contents := "#!/bin/sh\ncat <<'BOM'\n" + bom + "\nBOM\n"

	err := os.WriteFile(script, []byte(contents), 0755)
	require.NoError(t, err, "failed writing syft shim")

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestGenerateAttestationFiles(t *testing.T) {
	writeFakeSyft(t, minimalCycloneDXBOM)

	artifactPath, sha256 := writeTempArtifact(t, "binary bytes")

	meta := testMetadataObj()
	meta.BuildInfo.Attestations = true

	g := &Gomason{}

	err := g.GenerateAttestationFiles(meta, artifactPath, BuildTarget{Name: "linux/amd64"})
	require.NoError(t, err, "attestation generation must succeed with syft on PATH")

	// Provenance: envelope, binding, and predicate type all per the consumer contract.
	attPath := fmt.Sprintf("%s.att.json", artifactPath)
	attContents, err := os.ReadFile(attPath)
	require.NoError(t, err, "provenance file must be written beside the artifact")

	att := struct {
		Type          string `json:"_type"`
		PredicateType string `json:"predicateType"`
		Subject       []struct {
			Name   string            `json:"name"`
			Digest map[string]string `json:"digest"`
		} `json:"subject"`
		Predicate struct {
			BuildDefinition struct {
				ResolvedDependencies []struct {
					URI string `json:"uri"`
				} `json:"resolvedDependencies"`
			} `json:"buildDefinition"`
		} `json:"predicate"`
	}{}

	err = json.Unmarshal(attContents, &att)
	require.NoError(t, err, "provenance file must be valid JSON")

	assert.Equal(t, "https://in-toto.io/Statement/v1", att.Type, "provenance _type")
	assert.Equal(t, "https://slsa.dev/provenance/v1", att.PredicateType, "provenance predicateType")
	require.Len(t, att.Subject, 1, "provenance subject count")
	assert.Equal(t, "testproject", att.Subject[0].Name, "provenance subject name is the destination name")
	assert.Equal(t, sha256, att.Subject[0].Digest["sha256"], "provenance digest is the artifact checksum")
	require.Len(t, att.Predicate.BuildDefinition.ResolvedDependencies, 1, "provenance source dependency count")
	assert.Equal(t, "git+https://github.com/nikogura/testproject", att.Predicate.BuildDefinition.ResolvedDependencies[0].URI, "provenance source dependency is a git+https URI")

	// SBOM: same envelope, CycloneDX predicate, same binding.
	sbomPath := fmt.Sprintf("%s.sbom.json", artifactPath)
	sbomContents, err := os.ReadFile(sbomPath)
	require.NoError(t, err, "sbom file must be written beside the artifact")

	sbom := struct {
		Type          string `json:"_type"`
		PredicateType string `json:"predicateType"`
		Subject       []struct {
			Name   string            `json:"name"`
			Digest map[string]string `json:"digest"`
		} `json:"subject"`
		Predicate struct {
			BOMFormat string `json:"bomFormat"`
		} `json:"predicate"`
	}{}

	err = json.Unmarshal(sbomContents, &sbom)
	require.NoError(t, err, "sbom file must be valid JSON")

	assert.Equal(t, "https://in-toto.io/Statement/v1", sbom.Type, "sbom _type")
	assert.Equal(t, "https://cyclonedx.org/bom", sbom.PredicateType, "sbom predicateType")
	require.Len(t, sbom.Subject, 1, "sbom subject count")
	assert.Equal(t, "testproject", sbom.Subject[0].Name, "sbom subject name is the destination name")
	assert.Equal(t, sha256, sbom.Subject[0].Digest["sha256"], "sbom digest is the artifact checksum")
	assert.Equal(t, "CycloneDX", sbom.Predicate.BOMFormat, "sbom predicate carries the BOM")
}

func TestGenerateAttestationFilesSyftMissing(t *testing.T) {
	// Fail closed: attestations enabled but syft absent must error, and must not write
	// half the pair — a half-attested tool is refuse-bait on the consumer side.
	// An empty directory as the whole PATH: deterministic on any machine, including
	// ones with syft installed in /usr/bin.
	t.Setenv("PATH", t.TempDir())

	artifactPath, _ := writeTempArtifact(t, "binary bytes")

	meta := testMetadataObj()
	meta.BuildInfo.Attestations = true

	g := &Gomason{}

	err := g.GenerateAttestationFiles(meta, artifactPath, BuildTarget{Name: "linux/amd64"})
	require.Error(t, err, "missing syft with attestations enabled must fail the build")
	assert.Contains(t, err.Error(), "syft", "the error must name the missing tool")

	assert.NoFileExists(t, fmt.Sprintf("%s.att.json", artifactPath), "no provenance file on syft failure")
	assert.NoFileExists(t, fmt.Sprintf("%s.sbom.json", artifactPath), "no sbom file on syft failure")
}

func TestGenerateAttestationFilesNoPublishTarget(t *testing.T) {
	writeFakeSyft(t, minimalCycloneDXBOM)

	artifactPath, _ := writeTempArtifact(t, "binary bytes")

	meta := testMetadataObj()
	meta.BuildInfo.Attestations = true
	meta.PublishInfo = PublishInfo{}

	g := &Gomason{}

	err := g.GenerateAttestationFiles(meta, artifactPath, BuildTarget{Name: "linux/amd64"})
	require.Error(t, err, "attestations without a publish target must fail, not guess a subject name")

	assert.NoFileExists(t, fmt.Sprintf("%s.att.json", artifactPath), "no provenance file without a subject")
	assert.NoFileExists(t, fmt.Sprintf("%s.sbom.json", artifactPath), "no sbom file without a subject")
}

// generateTestGPGKey creates an isolated GPG home with a test key, following the same
// pattern as the sign tests, and returns the homedir.
func generateTestGPGKey(t *testing.T) (gpgHome string) {
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

// writeBuiltArtifact creates a fake built artifact in its own workdir, as gox would
// have dropped it there.
func writeBuiltArtifact(t *testing.T) (artifactPath string) {
	t.Helper()

	workDir := t.TempDir()
	artifactPath = filepath.Join(workDir, testfixtures.PublishSource)

	err := os.WriteFile(artifactPath, []byte("built bytes"), 0755)
	require.NoError(t, err, "failed writing fake artifact")

	return artifactPath
}

func TestProcessArtifactFileAttestations(t *testing.T) {
	writeFakeSyft(t, minimalCycloneDXBOM)

	gpgHome := generateTestGPGKey(t)

	artifactPath := writeBuiltArtifact(t)

	meta := testMetadataObj()
	meta.BuildInfo.Attestations = true
	meta.Options = map[string]interface{}{"homedir": gpgHome}

	cwd := t.TempDir()

	g := &Gomason{}

	// sign, collect, don't publish: the sign flow.
	err := g.processArtifactFile(meta, artifactPath, meta.BuildInfo.Targets[0], true, false, true, cwd)
	require.NoError(t, err, "sign flow with attestations must succeed")

	// Collect is a move: the workdir is drained into cwd, exactly as the binary and its
	// signature are. What lands in cwd is the full set of files a signed, attested
	// artifact carries.
	collectedSuffixes := []string{"", ".asc", ".att.json", ".att.json.asc", ".sbom.json", ".sbom.json.asc"}

	for _, suffix := range collectedSuffixes {
		assert.FileExists(t, filepath.Join(cwd, testfixtures.PublishSource+suffix), "artifact%s must be collected into cwd", suffix)
	}

	// And the attestation signatures verify against the signing key.
	for _, attFile := range []string{".att.json", ".sbom.json"} {
		ok, verifyErr := VerifyBinary(filepath.Join(cwd, testfixtures.PublishSource+attFile), meta)
		require.NoError(t, verifyErr, "verifying %s signature", attFile)
		assert.True(t, ok, "%s signature must verify against the signing key", attFile)
	}
}

func TestProcessArtifactFileWithoutAttestations(t *testing.T) {
	t.Parallel()

	// Opt-in only: attestations absent from metadata means the existing behavior,
	// byte for byte — no attestation files anywhere.
	gpgHome := generateTestGPGKey(t)

	artifactPath := writeBuiltArtifact(t)

	meta := testMetadataObj()
	meta.Options = map[string]interface{}{"homedir": gpgHome}

	cwd := t.TempDir()

	g := &Gomason{}

	err := g.processArtifactFile(meta, artifactPath, meta.BuildInfo.Targets[0], true, false, true, cwd)
	require.NoError(t, err, "sign flow without attestations must succeed")

	for _, suffix := range []string{".att.json", ".att.json.asc", ".sbom.json", ".sbom.json.asc"} {
		assert.NoFileExists(t, fmt.Sprintf("%s%s", artifactPath, suffix), "artifact%s must not exist without the opt-in", suffix)
		assert.NoFileExists(t, filepath.Join(cwd, testfixtures.PublishSource+suffix), "artifact%s must not be collected without the opt-in", suffix)
	}
}

func TestPublishFileAttestations(t *testing.T) {
	t.Parallel()

	var paths []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(http.StatusCreated)
	}))

	defer server.Close()

	artifactPath := writeBuiltArtifact(t)

	// The full set of files a signed, attested artifact carries.
	for _, suffix := range []string{".asc", ".att.json", ".att.json.asc", ".sbom.json", ".sbom.json.asc"} {
		err := os.WriteFile(fmt.Sprintf("%s%s", artifactPath, suffix), []byte("file contents"+suffix), 0644)
		require.NoError(t, err, "failed writing artifact side file %s", suffix)
	}

	meta := testMetadataObj()
	meta.Repository = fmt.Sprintf("%s/repo", server.URL)

	target := meta.PublishInfo.Targets[0]
	target.Signature = true
	target.Checksums = false
	target.Attestations = true

	meta.PublishInfo = PublishInfo{
		Targets:    []PublishTarget{target},
		TargetsMap: map[string]PublishTarget{testfixtures.PublishSource: target},
	}

	g := &Gomason{}

	err := g.PublishFile(meta, artifactPath)
	require.NoError(t, err, "publishing with attestations must succeed")

	expected := []string{
		"/repo/testproject/0.1.0/linux/amd64/testproject",
		"/repo/testproject/0.1.0/linux/amd64/testproject.asc",
		"/repo/testproject/0.1.0/linux/amd64/testproject.att.json",
		"/repo/testproject/0.1.0/linux/amd64/testproject.att.json.asc",
		"/repo/testproject/0.1.0/linux/amd64/testproject.sbom.json",
		"/repo/testproject/0.1.0/linux/amd64/testproject.sbom.json.asc",
	}

	for _, want := range expected {
		assert.Contains(t, paths, want, "publish must upload %s", want)
	}
}

func TestPublishFileAttestationsMissingFiles(t *testing.T) {
	t.Parallel()

	// Fail closed: the publish target asks for attestations, but the files were never
	// generated — the mismatch is an error, not a silent partial publish.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	defer server.Close()

	artifactPath := writeBuiltArtifact(t)

	meta := testMetadataObj()
	meta.Repository = fmt.Sprintf("%s/repo", server.URL)

	target := meta.PublishInfo.Targets[0]
	target.Signature = true
	target.Checksums = false
	target.Attestations = true

	meta.PublishInfo = PublishInfo{
		Targets:    []PublishTarget{target},
		TargetsMap: map[string]PublishTarget{testfixtures.PublishSource: target},
	}

	g := &Gomason{}

	err := g.PublishFile(meta, artifactPath)
	require.Error(t, err, "publishing attestations without the files must fail")
}

func TestBuildTargetForSource(t *testing.T) {
	t.Parallel()

	meta := testMetadataObj()

	t.Run("matching source maps to its build target", func(t *testing.T) {
		t.Parallel()

		target := BuildTargetForSource(meta, testfixtures.PublishSource)
		assert.Equal(t, testfixtures.BuildTargetName, target.Name, "artifact suffix maps to the build target")
	})

	t.Run("unmatched source maps to an empty target", func(t *testing.T) {
		t.Parallel()

		target := BuildTargetForSource(meta, "some_unrelated_artifact")
		assert.Empty(t, target.Name, "no match means no guess")
	})
}

func TestGitHTTPSUrlFromPackage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"plain module path",
			"github.com/nikogura/testproject",
			"git+https://github.com/nikogura/testproject",
		},
		{
			"major version suffix is not part of the repo url",
			"github.com/nikogura/gomason/v2",
			"git+https://github.com/nikogura/gomason",
		},
		{
			"non-github host",
			"example.com/tools/mytool",
			"git+https://example.com/tools/mytool",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, GitHTTPSUrlFromPackage(tc.input), "git+https URI from package path")
		})
	}
}
