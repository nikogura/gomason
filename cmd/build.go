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
	"github.com/nikogura/gomason/pkg/gomason"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
)

var buildSkipTests bool

// buildCmd represents the build command
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
		gm, err := gomason.NewGomason()
		if err != nil {
			log.Fatalf("error creating gomason object")
		}

		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current working directory: %s", err)
		}
		rootWorkDir, err := ioutil.TempDir("", "gomason")
		if err != nil {
			log.Fatalf("Failed to create temp dir: %s", err)
		}

		defer os.RemoveAll(rootWorkDir)

		meta, err := gomason.ReadMetadata(gomason.METADATA_FILENAME)
		if err != nil {
			log.Fatalf("couldn't read package information from metadata file: %s", err)
		}

		lang, err := gomason.GetByName(meta.GetLanguage())
		if err != nil {
			log.Fatalf("Invalid language: %v", err)
		}

		workDir, err := lang.CreateWorkDir(rootWorkDir)
		if err != nil {
			log.Fatalf("Failed to create ephemeral workDir: %s", err)
		}

		err = lang.Checkout(workDir, meta, branch)
		if err != nil {
			log.Fatalf("failed to checkout package %s at branch %s: %s", meta.Package, branch, err)
		}

		err = lang.Prep(workDir, meta)
		if err != nil {
			log.Fatalf("error running prep steps: %s", err)
		}

		if !buildSkipTests {
			err = lang.Test(workDir, meta.Package, testTimeout)
			if err != nil {
				log.Fatalf("error running go test: %s", err)
			}
		}

		err = lang.Build(workDir, meta, buildSkipTargets)
		if err != nil {
			log.Fatalf("build failed: %s", err)
		}

		err = gm.HandleArtifacts(meta, workDir, cwd, false, false, true, buildSkipTargets)
		if err != nil {
			log.Fatalf("signing failed: %s", err)
		}

		err = gm.HandleExtras(meta, workDir, cwd, false, false, true)
		if err != nil {
			log.Fatalf("Extra artifact processing failed: %s", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().BoolVarP(&buildSkipTests, "skiptests", "s", false, "Skip tests when building.")
}
