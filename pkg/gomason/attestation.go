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
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Attestation support: gomason as the producer of signed in-toto attestations that a
// distribution client (e.g. dbt, the Dynamic Binary Toolkit) verifies before running a
// tool. The contract is pinned from the consumer side:
//
//   - in-toto Statement v1 envelopes
//   - exactly one subject per statement: the name the artifact carries on the
//     repository, and the artifact's sha256
//   - predicateType https://slsa.dev/provenance/v1 for .att.json,
//     https://cyclonedx.org/bom for .sbom.json
//   - files published as <destination>.att.json, <destination>.att.json.asc,
//     <destination>.sbom.json, <destination>.sbom.json.asc — or none at all
//
// The subject name is deliberately derived from the publish destination, not from the
// module or artifact filename: rebranded distributions rename the binary (an
// "acme"-branded dbt publishes as acme, while the module is still
// dynamic-binary-toolkit), and the client verifies the subject against the name the
// binary actually carries. The destination is the only place that name is known.

// Contract constants. These must match the consumer's pinned values.
const (
	attestationStatementType      = "https://in-toto.io/Statement/v1"
	attestationPredicateSLSA      = "https://slsa.dev/provenance/v1"
	attestationPredicateCycloneDX = "https://cyclonedx.org/bom"

	// gomasonBuildType identifies a gomason clean-workspace build in SLSA provenance.
	gomasonBuildType = "https://gomason.nikogura.com/build/v1"
)

// attestationStatement is an in-toto Statement v1 as gomason emits it. The predicate is
// carried as raw JSON: gomason generates it, and interpretation belongs to policy
// tooling, not to the producer or the distribution client.
type attestationStatement struct {
	Type          string               `json:"_type"`
	PredicateType string               `json:"predicateType"`
	Subject       []attestationSubject `json:"subject"`
	Predicate     json.RawMessage      `json:"predicate"`
}

// attestationSubject is the single artifact a statement binds its predicate to.
type attestationSubject struct {
	Name   string            `json:"name"`
	Digest map[string]string `json:"digest"`
}

// provenanceInput is everything a SLSA provenance statement needs, resolved by the
// caller. builder is the signing identity (the trust-the-person model: the builder is
// the person whose key signs, not a platform); when empty, the signing email from
// metadata is used.
type provenanceInput struct {
	meta      Metadata
	target    string
	subject   string
	sha256    string
	sourceURI string
	commit    string
	builder   string
}

// generateProvenanceStatement builds the SLSA v1.0 provenance statement for one built
// artifact. The statement claims what gomason actually knows: the package, version, and
// target from metadata; the source URI and commit of the checkout; and the identity
// whose key signs the artifacts. It does not claim isolation or a hosted platform —
// gomason builds are local builds signed by a person, and the statement says so by
// saying nothing more.
//
// Marshaling a locally-constructed struct cannot fail in a healthy world, but if it
// ever does, the failure is returned — an empty envelope must never be written to
// disk and signed as if it were a real statement.
func generateProvenanceStatement(input provenanceInput) (stmt attestationStatement, err error) {
	builder := input.builder
	if builder == "" {
		builder = input.meta.SignInfo.Email
	}

	predicate := provenancePredicate{
		BuildDefinition: buildDefinition{
			BuildType: gomasonBuildType,
			ExternalParameters: externalParameters{
				Package: input.meta.Package,
				Version: input.meta.Version,
				Target:  input.target,
			},
			ResolvedDeps: resolvedDependencies(input.sourceURI, input.commit),
		},
		RunDetails: runDetails{
			Builder: builderIdentity{
				ID: builder,
			},
			Metadata: runMetadata{
				InvocationID: generateInvocationID(),
			},
		},
	}

	predicateBytes, marshalErr := json.Marshal(predicate)
	if marshalErr != nil {
		err = errors.Wrap(marshalErr, "failed to marshal provenance predicate")
		return stmt, err
	}

	stmt = attestationStatement{
		Type:          attestationStatementType,
		PredicateType: attestationPredicateSLSA,
		Subject:       []attestationSubject{subjectOf(input.subject, input.sha256)},
		Predicate:     predicateBytes,
	}

	return stmt, err
}

// generateSBOMStatement wraps a bare CycloneDX BOM (as syft emits) into an in-toto
// Statement bound to the artifact. A bare BOM carries no subject digest — it is a
// forgeable text file; the statement envelope is what binds it to these exact bytes.
func generateSBOMStatement(subject string, sha256 string, bom json.RawMessage) (stmt attestationStatement) {
	stmt = attestationStatement{
		Type:          attestationStatementType,
		PredicateType: attestationPredicateCycloneDX,
		Subject:       []attestationSubject{subjectOf(subject, sha256)},
		Predicate:     bom,
	}

	return stmt
}

// subjectOf builds the single-subject binding for an artifact.
func subjectOf(name string, sha256 string) (subject attestationSubject) {
	subject = attestationSubject{
		Name:   name,
		Digest: map[string]string{"sha256": sha256},
	}

	return subject
}

// resolvedDependencies builds the source dependency entry. A 40-char commit is sha1, a
// 64-char commit is sha256; no commit means the URI is carried without a digest rather
// than claiming one that isn't known.
func resolvedDependencies(sourceURI string, commit string) (deps []provenanceDependency) {
	dep := provenanceDependency{
		URI: sourceURI,
	}

	if commit != "" {
		algorithm := "sha1"
		if len(commit) == 64 {
			algorithm = "sha256"
		}

		dep.Digest = map[string]string{algorithm: commit}
	}

	deps = append(deps, dep)

	return deps
}

// generateInvocationID mints a random id for this build run.
func generateInvocationID() (id string) {
	bytes := make([]byte, 16)

	// A failed read leaves zero bytes: the id degrades to a deterministic value
	// rather than failing the build over a nonce.
	_, _ = rand.Read(bytes)

	id = hex.EncodeToString(bytes)

	return id
}

// attestationSubjectName resolves the name the artifact carries on the repository: the
// final path segment of its publish destination, after template resolution. This is the
// name a distribution client verifies the subject against, and for rebranded
// distributions it is the only correct source — the branded name lives in the
// destination, not in the module or the local artifact filename.
func attestationSubjectName(meta Metadata, artifactBase string) (subject string, err error) {
	target, ok := meta.PublishInfo.TargetsMap[artifactBase]
	if !ok {
		err = fmt.Errorf("no publish target for artifact %q: attestations need a destination to name the subject", artifactBase)
		return subject, err
	}

	parsedDestination, parseErr := ParseTemplateForMetadata(target.Destination, meta)
	if parseErr != nil {
		err = errors.Wrapf(parseErr, "failed to parse destination %q", target.Destination)
		return subject, err
	}

	segments := strings.Split(strings.TrimSuffix(parsedDestination, "/"), "/")
	subject = segments[len(segments)-1]

	if subject == "" {
		err = fmt.Errorf("publish destination %q has no file name to use as an attestation subject", target.Destination)
		return subject, err
	}

	return subject, err
}

// provenancePredicate is the SLSA v1.0 provenance predicate gomason emits.
type provenancePredicate struct {
	BuildDefinition buildDefinition `json:"buildDefinition"`
	RunDetails      runDetails      `json:"runDetails"`
}

// buildDefinition captures how the artifact was built.
type buildDefinition struct {
	BuildType          string                 `json:"buildType"`
	ExternalParameters externalParameters     `json:"externalParameters"`
	ResolvedDeps       []provenanceDependency `json:"resolvedDependencies"`
}

// externalParameters is the build input: which package, at what version, for which target.
type externalParameters struct {
	Package string `json:"package"`
	Version string `json:"version"`
	Target  string `json:"target"`
}

// provenanceDependency is one resolved input to the build.
type provenanceDependency struct {
	URI    string            `json:"uri"`
	Digest map[string]string `json:"digest,omitempty"`
}

// runDetails captures who ran the build.
type runDetails struct {
	Builder  builderIdentity `json:"builder"`
	Metadata runMetadata     `json:"metadata"`
}

// builderIdentity is the person or platform that ran the build. For gomason this is the
// signing identity: the builder is the signer.
type builderIdentity struct {
	ID string `json:"id"`
}

// runMetadata carries per-run identifiers.
type runMetadata struct {
	InvocationID string `json:"invocationId"`
}

// GenerateAttestationFiles generates the SLSA provenance attestation (<artifact>.att.json)
// and the CycloneDX SBOM attestation (<artifact>.sbom.json) for a built artifact, writing
// them beside the artifact. Files are written only after everything needed for both has
// succeeded — the pair is never half-published, because the consumer treats a
// half-attested tool as a botched publish and refuses it.
func (g *Gomason) GenerateAttestationFiles(meta Metadata, artifactPath string, target BuildTarget) (err error) {
	var subject string
	var sha256 string
	var bom json.RawMessage
	var provenanceBytes []byte
	var sbomBytes []byte

	artifactBase := filepath.Base(artifactPath)

	subject, err = attestationSubjectName(meta, artifactBase)
	if err != nil {
		return err
	}

	sha256, err = FileSha256(artifactPath)
	if err != nil {
		err = errors.Wrapf(err, "failed to checksum %s", artifactPath)
		return err
	}

	// Syft first: it is the one step that can fail independently, and it must fail
	// before any file is written.
	bom, err = runSyft(artifactPath)
	if err != nil {
		return err
	}

	logrus.Debugf("Generating attestations for %s (subject %q)", artifactPath, subject)

	provenance, provenanceErr := generateProvenanceStatement(provenanceInput{
		meta:      meta,
		target:    target.Name,
		subject:   subject,
		sha256:    sha256,
		sourceURI: GitHTTPSUrlFromPackage(meta.Package),
		commit:    gitCommit(filepath.Dir(artifactPath)),
		builder:   g.signingEntity(meta),
	})
	if provenanceErr != nil {
		err = errors.Wrap(provenanceErr, "failed to generate provenance statement")
		return err
	}

	provenanceBytes, err = json.MarshalIndent(provenance, "", "  ")
	if err != nil {
		err = errors.Wrap(err, "failed to marshal provenance statement")
		return err
	}

	sbom := generateSBOMStatement(subject, sha256, bom)

	sbomBytes, err = json.MarshalIndent(sbom, "", "  ")
	if err != nil {
		err = errors.Wrap(err, "failed to marshal sbom statement")
		return err
	}

	attPath := fmt.Sprintf("%s.att.json", artifactPath)

	err = os.WriteFile(attPath, provenanceBytes, 0644)
	if err != nil {
		err = errors.Wrapf(err, "failed to write %s", attPath)
		return err
	}

	sbomPath := fmt.Sprintf("%s.sbom.json", artifactPath)

	err = os.WriteFile(sbomPath, sbomBytes, 0644)
	if err != nil {
		err = errors.Wrapf(err, "failed to write %s", sbomPath)
		return err
	}

	return err
}

// runSyft shells out to syft for a bare CycloneDX BOM of the artifact. Syft does the
// heavy lifting — gomason deliberately does not reimplement package identification.
// Fail-closed: absent syft, or output that is not a CycloneDX BOM, is an error.
func runSyft(artifactPath string) (bom json.RawMessage, err error) {
	var syftPath string

	syftPath, err = exec.LookPath("syft")
	if err != nil {
		err = errors.Wrap(err, "syft is required to generate attestations (metadata has attestations enabled) but was not found in PATH. Install syft: https://github.com/anchore/syft")
		return bom, err
	}

	logrus.Debugf("Running %s scan %s -o cyclonedx-json", syftPath, artifactPath)

	var out bytes.Buffer

	cmd := exec.CommandContext(context.Background(), syftPath, "scan", artifactPath, "-o", "cyclonedx-json")
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	runErr := cmd.Run()
	if runErr != nil {
		err = errors.Wrapf(runErr, "syft failed scanning %s", artifactPath)
		return bom, err
	}

	bom = json.RawMessage(out.Bytes())

	// The output must actually be a CycloneDX BOM — a shim, a version change, or a
	// misuse must not be wrapped into an attestation that claims otherwise.
	bomCheck := struct {
		BOMFormat string `json:"bomFormat"`
	}{}

	checkErr := json.Unmarshal(bom, &bomCheck)
	if checkErr != nil {
		err = errors.Wrap(checkErr, "syft output was not valid JSON")
		return bom, err
	}

	if bomCheck.BOMFormat != "CycloneDX" {
		err = fmt.Errorf("syft output was not a CycloneDX BOM (bomFormat %q)", bomCheck.BOMFormat)
		return bom, err
	}

	return bom, err
}

// gitCommit returns the HEAD commit of a git checkout, or an empty string when the
// directory is not a usable checkout. Best-effort: provenance carries the commit when it
// is known and claims nothing when it is not.
func gitCommit(dir string) (commit string) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return commit
	}

	var out bytes.Buffer

	cmd := exec.CommandContext(context.Background(), gitPath, "-C", dir, "rev-parse", "HEAD")
	cmd.Stdout = &out

	runErr := cmd.Run()
	if runErr != nil {
		logrus.Debugf("no git commit available in %s: %s", dir, runErr)

		return commit
	}

	commit = strings.TrimSpace(out.String())

	return commit
}

// signingEntity resolves the signing identity: the metadata's signing email, overridden
// by the user config in ~/.gomason. The same resolution signing uses — the attestation's
// builder is the person whose key signs the artifacts.
func (g *Gomason) signingEntity(meta Metadata) (entity string) {
	entity = meta.SignInfo.Email

	if g.Config.User.Email != "" {
		entity = g.Config.User.Email
	}

	return entity
}

// BuildTargetForSource maps a built artifact name back to its build target by the
// os/arch suffix gox gave it (testproject_linux_amd64 -> linux/amd64). Empty when
// nothing matches — provenance then carries an empty target parameter rather than a
// guess.
func BuildTargetForSource(meta Metadata, source string) (target BuildTarget) {
	for _, t := range meta.BuildInfo.Targets {
		if strings.HasSuffix(source, fmt.Sprintf("_%s", strings.ReplaceAll(t.Name, "/", "_"))) {
			target = t

			return target
		}
	}

	return target
}
