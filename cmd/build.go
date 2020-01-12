// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current working directory: %s", err)
		}
		rootWorkDir, err := ioutil.TempDir("", "gomason")
		if err != nil {
			log.Fatalf("Failed to create temp dir: %s", err)
		}

		log.Printf("[DEBUG] Created temp dir %s", rootWorkDir)

		defer os.RemoveAll(rootWorkDir)

		meta, err := gomason.ReadMetadata("metadata.json")
		if err != nil {
			log.Fatalf("couldn't read package information from metadata.json: %s", err)
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
			err = lang.Test(workDir, meta.Package)
			if err != nil {
				log.Fatalf("error running go test: %s", err)
			}

			fmt.Print("Tests Succeeded!\n\n")
		}

		err = lang.Build(workDir, meta)
		if err != nil {
			log.Fatalf("build failed: %s", err)
		}

		fmt.Print("Build Succeeded!\n\n")

		err = gomason.HandleArtifacts(meta, workDir, cwd, false, false, true)
		if err != nil {
			log.Fatalf("signing failed: %s", err)
		}

		err = gomason.HandleExtras(meta, workDir, cwd, false, false)
		if err != nil {
			log.Fatalf("Extra artifact processing failed: %s", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// buildCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	buildCmd.Flags().BoolVarP(&buildSkipTests, "skiptests", "s", false, "Skip tests when building.")
}
