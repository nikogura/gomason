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

// signCmd represents the sign command.
//
//nolint:gochecknoglobals // Cobra boilerplate
var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign your binaries after building them.",
	Long: `
Sign your binaries after building them.

Artists sign their work, you should too.

Signing sorta implies something to sign, which in turn, implies that it built, which means it tested successfully.  What I'm getting at is this command will run 'test', 'build', and then it will 'sign'.
`,
	Run: func(cmd *cobra.Command, args []string) {
		rootWorkDir, runErr := runSign()

		cleanupErr := cleanupWorkDir(rootWorkDir)

		if runErr != nil {
			log.Fatalf("%s", runErr)
		}

		if cleanupErr != nil {
			log.Fatalf("%s", cleanupErr)
		}
	},
}

// runSign tests, builds, and signs the project in a clean, ephemeral work dir.
// It returns the root work dir it created so the caller can clean it up on
// every path, including the failure paths.
func runSign() (rootWorkDir string, err error) {
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

	var workDir string

	rootWorkDir, workDir, err = prepareWorkspace(lang, meta)
	if err != nil {
		return rootWorkDir, err
	}

	err = lang.Prep(workDir, meta, local)
	if err != nil {
		err = fmt.Errorf("error running prep steps: %w", err)

		return rootWorkDir, err
	}

	err = lang.Test(workDir, meta.Package, testTimeout, local)
	if err != nil {
		err = fmt.Errorf("error running go test: %w", err)

		return rootWorkDir, err
	}

	log.Printf("Tests Succeeded!\n\n")

	err = lang.Build(workDir, meta, buildSkipTargets, local)
	if err != nil {
		err = fmt.Errorf("build failed: %w", err)

		return rootWorkDir, err
	}

	err = gm.HandleArtifacts(meta, workDir, cwd, true, false, true, buildSkipTargets, local)
	if err != nil {
		err = fmt.Errorf("signing failed: %w", err)

		return rootWorkDir, err
	}

	err = gm.HandleExtras(meta, workDir, cwd, true, false, true, local)
	if err != nil {
		err = fmt.Errorf("extra artifact processing failed: %w", err)

		return rootWorkDir, err
	}

	return rootWorkDir, err
}

//nolint:gochecknoinits // Cobra boilerplate
func init() {
	rootCmd.AddCommand(signCmd)
}
