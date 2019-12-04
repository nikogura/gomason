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
	"io/ioutil"
	"log"
	"os"

	"github.com/nikogura/gomason/pkg/gomason"
	"github.com/nikogura/gomason/pkg/gomason/languages"
	"github.com/spf13/cobra"
)

var pubSkipTests bool

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Test, build, sign and publish your code",
	Long: `
Test, build, sign and publish your code.

Publish will upload your binaries to wherever it is you've configured them to go in whatever way you like.  The detached signatures will likewise be uploaded.
`,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current working directory: %s", err)
		}
		rootWorkDir, err := ioutil.TempDir("", "gomason")
		if err != nil {
			log.Fatalf("Failed to create temp dir: %s", err)
		}
		defer os.RemoveAll(rootWorkDir)

		if verbose {
			log.Printf("Created temp dir %s", rootWorkDir)
		}

		meta, err := gomason.ReadMetadata("metadata.json")
		if err != nil {
			log.Fatalf("failed to read metadata: %s", err)
		}

		lang, err := languages.GetByName(meta.GetLanguage())
		if err != nil {
			log.Fatalf("Invalid language: %v", err)
		}

		workDir, err := lang.CreateWorkDir(rootWorkDir)
		if err != nil {
			log.Fatalf("Failed to create ephemeral working directory: %s", err)
		}

		err = lang.Checkout(workDir, meta, branch, verbose)
		if err != nil {
			log.Fatalf("failed to checkout package %s at branch %s: %s", meta.Package, branch, err)
		}

		err = lang.Prep(workDir, meta, verbose)
		if err != nil {
			log.Fatalf("error running prep steps: %s", err)
		}

		if !pubSkipTests {
			err = lang.Test(workDir, meta.Package, verbose)
			if err != nil {
				log.Fatalf("error running go test: %s", err)
			}

			log.Printf("Tests Succeeded!\n\n")
		}

		err = lang.Build(workDir, meta, branch, verbose)
		if err != nil {
			log.Fatalf("build failed: %s", err)
		}

		log.Printf("Build Succeeded!\n\n")

		if meta.PublishInfo.SkipSigning {
			if verbose {
				log.Printf("Skipping signing due to 'skip-signing': true in metadata.json")
			}
			err = gomason.HandleArtifacts(meta, workDir, cwd, false, true, false, verbose)
			if err != nil {
				log.Fatalf("post-build processing failed: %s", err)
			}

			err = gomason.HandleExtras(meta, workDir, cwd, false, true, verbose)
			if err != nil {
				log.Fatalf("Extra artifact processing failed: %s", err)
			}

		} else {
			err = gomason.HandleArtifacts(meta, workDir, cwd, true, true, false, verbose)
			if err != nil {
				log.Fatalf("post-build processing failed: %s", err)
			}

			err = gomason.HandleExtras(meta, workDir, cwd, true, true, verbose)
			if err != nil {
				log.Fatalf("Extra artifact processing failed: %s", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// publishCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// publishCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	publishCmd.Flags().BoolVarP(&pubSkipTests, "skiptests", "s", false, "Skip tests when publishing.")
}
