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
	"io/ioutil"
	"log"
	"os"

	"github.com/nikogura/gomason/pkg/gomason"
	"github.com/spf13/cobra"
)

var pubSkipTests bool
var pubSkipBuild bool

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Test, build, sign and publish your code",
	Long: `
Test, build, sign and publish your code.

Publish will upload your binaries to wherever it is you've configured them to go in whatever way you like.  The detached signatures will likewise be uploaded.
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

		log.Printf("[DEBUG] Created temp dir %s", rootWorkDir)

		meta, err := gomason.ReadMetadata(gomason.METADATA_FILENAME)
		if err != nil {
			log.Fatalf("failed to read metadata: %s", err)
		}

		lang, err := gomason.GetByName(meta.GetLanguage())
		if err != nil {
			log.Fatalf("Invalid language: %v", err)
		}

		workDir, err := lang.CreateWorkDir(rootWorkDir)
		if err != nil {
			log.Fatalf("Failed to create ephemeral working directory: %s", err)
		}

		// Totally skip building, and just do signing and uploading
		if pubSkipBuild {
			workDir, err = os.Getwd()
			if err != nil {
				log.Fatalf("Failed getting current working directory.")
			}

			fmt.Printf("Workdir is %s\n", workDir)

			for _, t := range meta.PublishInfo.Targets {
				if meta.PublishInfo.SkipSigning {
					err = gm.PublishFile(meta, t.Source)
					if err != nil {
						log.Fatalf("Failed to publish %s: %s", t.Source, err)
					}

				} else {
					err = gm.SignBinary(meta, t.Source)
					if err != nil {
						log.Fatalf("Failed to sign %s: %s", t.Source, err)
					}

					err = gm.PublishFile(meta, t.Source)
					if err != nil {
						log.Fatalf("Failed to publish %s: %s", t.Source, err)
					}
				}
			}
		} else {
			err = lang.Checkout(workDir, meta, branch)
			if err != nil {
				log.Fatalf("failed to checkout package %s at branch %s: %s", meta.Package, branch, err)
			}

			err = lang.Prep(workDir, meta)
			if err != nil {
				log.Fatalf("error running prep steps: %s", err)
			}

			if !pubSkipTests {
				err = lang.Test(workDir, meta.Package)
				if err != nil {
					log.Fatalf("error running go test: %s", err)
				}

				fmt.Print("Tests Succeeded!\n\n")
			}

			err = lang.Build(workDir, meta, buildSkipTargets)
			if err != nil {
				log.Fatalf("build failed: %s", err)
			}

			fmt.Print("Build Succeeded!\n\n")

			if meta.PublishInfo.SkipSigning {
				log.Printf("[DEBUG] Skipping signing due to 'skip-signing': true in metadata file")
				err = gm.HandleArtifacts(meta, workDir, cwd, false, true, false, buildSkipTargets)
				if err != nil {
					log.Fatalf("post-build processing failed: %s", err)
				}

				err = gm.HandleExtras(meta, workDir, cwd, false, true)
				if err != nil {
					log.Fatalf("Extra artifact processing failed: %s", err)
				}

			} else {
				err = gm.HandleArtifacts(meta, workDir, cwd, true, true, false, buildSkipTargets)
				if err != nil {
					log.Fatalf("post-build processing failed: %s", err)
				}

				err = gm.HandleExtras(meta, workDir, cwd, true, true)
				if err != nil {
					log.Fatalf("Extra artifact processing failed: %s", err)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)

	publishCmd.Flags().BoolVarP(&pubSkipTests, "skiptests", "s", false, "Skip tests when publishing.")
	publishCmd.Flags().BoolVarP(&pubSkipBuild, "skipbuild", "", false, "Skip build altogether and only publish.")
}
