// Copyright © 2017 Nik Ogura <nik.ogura@gmail.com>
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
	"log"
	"os"

	"github.com/nikogura/gomason/v2/pkg/gomason"
	"github.com/spf13/cobra"
)

//nolint:gochecknoglobals // Cobra boilerplate
var pubSkipTests bool

//nolint:gochecknoglobals // Cobra boilerplate
var pubSkipBuild bool

// publishCmd represents the publish command.
//
//nolint:gochecknoglobals // Cobra boilerplate
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Test, build, sign and publish your code",
	Long: `
Test, build, sign and publish your code.

Publish will upload your binaries to wherever it is you've configured them to go in whatever way you like.  The detached signatures will likewise be uploaded.
`,
	Run: func(cmd *cobra.Command, args []string) {
		rootWorkDir, runErr := runPublish()

		cleanupErr := cleanupWorkDir(rootWorkDir)

		if runErr != nil {
			log.Fatalf("%s", runErr)
		}

		if cleanupErr != nil {
			log.Fatalf("%s", cleanupErr)
		}
	},
}

// runPublish tests, builds, signs, and publishes the project in a clean,
// ephemeral work dir.  It returns the root work dir it created so the caller
// can clean it up on every path, including the failure paths.
func runPublish() (rootWorkDir string, err error) {
	var gm *gomason.Gomason

	gm, err = gomason.NewGomason()
	if err != nil {
		err = fmt.Errorf("error creating gomason object: %w", err)

		return rootWorkDir, err
	}

	var cwd string

	cwd, err = os.Getwd()
	if err != nil {
		err = fmt.Errorf("failed to get current working directory: %w", err)

		return rootWorkDir, err
	}

	var meta gomason.Metadata

	meta, err = gomason.ReadMetadata(gomason.MetadataFilename)
	if err != nil {
		err = fmt.Errorf("failed to read metadata: %w", err)

		return rootWorkDir, err
	}

	var lang gomason.Language

	lang, err = gomason.GetByName(meta.GetLanguage())
	if err != nil {
		err = fmt.Errorf("invalid language: %w", err)

		return rootWorkDir, err
	}

	// Totally skip building, and just do signing and uploading.
	if pubSkipBuild {
		err = publishPrebuilt(gm, meta)

		return rootWorkDir, err
	}

	var workDir string

	rootWorkDir, err = createEphemeralWorkDir()
	if err != nil {
		return rootWorkDir, err
	}

	workDir, err = lang.CreateWorkDir(rootWorkDir)
	if err != nil {
		err = fmt.Errorf("failed to create ephemeral working directory: %w", err)

		return rootWorkDir, err
	}

	err = buildAndPublish(gm, lang, meta, workDir, cwd)

	return rootWorkDir, err
}

// publishPrebuilt signs and uploads artifacts that were built elsewhere,
// without building anything itself. Attestations, when enabled, are generated here too —
// from the prebuilt binaries, exactly as the built path generates them from its own —
// so a publish target that asks for attestations is satisfiable either way.
func publishPrebuilt(gm *gomason.Gomason, meta gomason.Metadata) (err error) {
	for _, t := range meta.PublishInfo.Targets {
		if !meta.PublishInfo.SkipSigning {
			err = gm.SignBinary(meta, t.Source)
			if err != nil {
				err = fmt.Errorf("failed to sign %s: %w", t.Source, err)

				return err
			}

			if meta.BuildInfo.Attestations {
				err = attestPrebuilt(gm, meta, t.Source)
				if err != nil {
					err = fmt.Errorf("failed to generate attestations for %s: %w", t.Source, err)

					return err
				}
			}
		}

		err = gm.PublishFile(meta, t.Source)
		if err != nil {
			err = fmt.Errorf("failed to publish %s: %w", t.Source, err)

			return err
		}
	}

	return err
}

// attestPrebuilt generates and signs attestation files for a prebuilt artifact. It
// refuses when the artifact's name maps to no build target: provenance for it would
// have to carry a blank target field, which is a quiet lie about how the binary was
// built.
func attestPrebuilt(gm *gomason.Gomason, meta gomason.Metadata, source string) (err error) {
	buildTarget := gomason.BuildTargetForSource(meta, source)
	if buildTarget.Name == "" {
		err = fmt.Errorf("no build target matches the name %q, so provenance would carry a blank target field", source)

		return err
	}

	err = gm.HandleAttestations(meta, source, buildTarget)
	if err != nil {
		err = fmt.Errorf("attestation generation failed: %w", err)

		return err
	}

	return err
}

// buildAndPublish runs the full pipeline: checkout, prep, test, build, then
// sign and upload whatever came out of it.
func buildAndPublish(gm *gomason.Gomason, lang gomason.Language, meta gomason.Metadata, workDir, cwd string) (err error) {
	if !local {
		err = lang.Checkout(workDir, meta, branch)
		if err != nil {
			err = fmt.Errorf("failed to checkout package %s at branch %s: %w", meta.Package, branch, err)

			return err
		}
	}

	err = lang.Prep(workDir, meta, local)
	if err != nil {
		err = fmt.Errorf("error running prep steps: %w", err)

		return err
	}

	if !pubSkipTests {
		err = lang.Test(workDir, meta.Package, testTimeout, local)
		if err != nil {
			err = fmt.Errorf("error running go test: %w", err)

			return err
		}
	}

	err = lang.Build(workDir, meta, buildSkipTargets, local)
	if err != nil {
		err = fmt.Errorf("build failed: %w", err)

		return err
	}

	// A signed publish signs the artifacts; an unsigned one skips straight to
	// upload.  Everything else about the two paths is identical: publishing is
	// on, and collecting artifacts locally is off.
	sign := !meta.PublishInfo.SkipSigning
	if !sign {
		log.Printf("[DEBUG] Skipping signing due to 'skip-signing': true in metadata file")
	}

	const (
		publishArtifacts = true
		collectArtifacts = false
	)

	err = gm.HandleArtifacts(meta, workDir, cwd, sign, publishArtifacts, collectArtifacts, buildSkipTargets, local)
	if err != nil {
		err = fmt.Errorf("post-build processing failed: %w", err)

		return err
	}

	err = gm.HandleExtras(meta, workDir, cwd, sign, publishArtifacts, collectArtifacts, local)
	if err != nil {
		err = fmt.Errorf("extra artifact processing failed: %w", err)

		return err
	}

	return err
}

//nolint:gochecknoinits // Cobra boilerplate
func init() {
	rootCmd.AddCommand(publishCmd)

	publishCmd.Flags().BoolVarP(&pubSkipTests, "skiptests", "s", false, "Skip tests when publishing.")
	publishCmd.Flags().BoolVarP(&pubSkipBuild, "skipbuild", "", false, "Skip build altogether and only publish.")
}
