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
	"github.com/nikogura/gomason/internal/app/gomason"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
)

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
		workDir, err := ioutil.TempDir("", "gomason")
		if err != nil {
			log.Fatalf("Failed to create temp dir: %s", err)
		}

		if verbose {
			log.Printf("Created temp dir %s", workDir)
		}

		defer os.RemoveAll(workDir)

		gopath, err := gomason.CreateGoPath(workDir)
		if err != nil {
			log.Fatalf("Failed to create ephemeral GOPATH: %s", err)
		}

		meta, err := gomason.ReadMetadata("metadata.json")

		err = gomason.GovendorInstall(gopath, verbose)
		if err != nil {
			log.Fatalf("Failed to install Govendor: %s", err)
		}

		if err != nil {
			log.Fatalf("couldn't read package information from metadata.json: %s", err)

		}

		err = gomason.Checkout(gopath, meta, branch, verbose)
		if err != nil {
			log.Fatalf("failed to checkout package %s at branch %s: %s", meta.Package, branch, err)
		}

		err = gomason.GovendorSync(gopath, meta, verbose)
		if err != nil {
			log.Fatalf("error running govendor sync: %s", err)
		}

		err = gomason.GoTest(gopath, meta.Package, verbose)
		if err != nil {
			log.Fatalf("error running go test: %s", err)
		}

		log.Printf("Tests Succeeded!\n\n")

		err = gomason.Build(gopath, meta, branch, verbose)
		if err != nil {
			log.Fatalf("build failed: %s", err)
		}

		log.Printf("Build Succeeded!\n\n")

		if meta.PublishInfo.SkipSigning {
			if verbose {
				log.Printf("Skipping signing due to 'skip-signing': true in metadata.json")
			}
			err = gomason.PublishBuildTargets(meta, gopath, cwd, false, true, false, verbose)
			if err != nil {
				log.Fatalf("post-build processing failed: %s", err)
			}

			err = gomason.PublishBuildExtras(meta, gopath, cwd, false, true, verbose)
			if err != nil {
				log.Fatalf("Extra artifact processing failed: %s", err)
			}

		} else {
			err = gomason.PublishBuildTargets(meta, gopath, cwd, true, true, false, verbose)
			if err != nil {
				log.Fatalf("post-build processing failed: %s", err)
			}

			err = gomason.PublishBuildExtras(meta, gopath, cwd, true, true, verbose)
			if err != nil {
				log.Fatalf("Extra artifact processing failed: %s", err)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(publishCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// publishCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// publishCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
