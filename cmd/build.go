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
var buildSkipTests bool

// buildCmd represents the build command.
//
//nolint:gochecknoglobals // Cobra boilerplate
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build your code in a clean environment.",
	Long: `
Build your code in a clean environment.

Includes 'test'.  It aint gonna build if the tests don't pass.

You could run 'test' separately, but 'build' is nice enough to do it for you.

Binaries are dropped into the current working directory.
`,
	Run: func(cmd *cobra.Command, args []string) {
		rootWorkDir, runErr := runBuild()

		cleanupErr := cleanupWorkDir(rootWorkDir)

		if runErr != nil {
			log.Fatalf("%s", runErr)
		}

		if cleanupErr != nil {
			log.Fatalf("%s", cleanupErr)
		}
	},
}

// runBuild tests and builds the project.  Unless --local was given, it works
// in a clean, ephemeral work dir, and returns the root of that dir so the
// caller can clean it up on every path, including the failure paths.  A local
// build works out of the current directory and returns an empty path, which
// cleanupWorkDir treats as nothing to do.
func runBuild() (rootWorkDir string, err error) {
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
		err = fmt.Errorf("couldn't read package information from metadata file: %w", err)

		return rootWorkDir, err
	}

	var lang gomason.Language

	lang, err = gomason.GetByName(meta.GetLanguage())
	if err != nil {
		err = fmt.Errorf("invalid language: %w", err)

		return rootWorkDir, err
	}

	// A local build uses whatever is checked out here; otherwise build clean.
	workDir := cwd

	if !local {
		rootWorkDir, workDir, err = prepareWorkspace(lang, meta)
		if err != nil {
			return rootWorkDir, err
		}
	}

	err = lang.Prep(workDir, meta, local)
	if err != nil {
		err = fmt.Errorf("error running prep steps: %w", err)

		return rootWorkDir, err
	}

	if !buildSkipTests {
		err = lang.Test(workDir, meta.Package, testTimeout, local)
		if err != nil {
			err = fmt.Errorf("error running go test: %w", err)

			return rootWorkDir, err
		}
	}

	err = lang.Build(workDir, meta, buildSkipTargets, local)
	if err != nil {
		err = fmt.Errorf("build failed: %w", err)

		return rootWorkDir, err
	}

	err = gm.HandleArtifacts(meta, workDir, cwd, false, false, true, buildSkipTargets, local)
	if err != nil {
		err = fmt.Errorf("signing failed: %w", err)

		return rootWorkDir, err
	}

	err = gm.HandleExtras(meta, workDir, cwd, false, false, true, local)
	if err != nil {
		err = fmt.Errorf("extra artifact processing failed: %w", err)

		return rootWorkDir, err
	}

	return rootWorkDir, err
}

//nolint:gochecknoinits // Cobra boilerplate
func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().BoolVarP(&buildSkipTests, "skiptests", "s", false, "Skip tests when building.")
}
