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
	"github.com/spf13/cobra"
)

// signCmd represents the sign command
var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign your binaries after building them.",
	Long: `
Sign your binaries after building them.

Artists sign their work, you should too.

Signing sorta implies something to sign, which in turn, implies that it built, which means it tested successfully.  What I'm getting at is this command will run 'test', 'build', and then it will 'sign'.
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
			log.Fatalf("failed to read metadata: %s", err)
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

		err = lang.Test(workDir, meta.Package, testTimeout, local)
		if err != nil {
			log.Fatalf("error running go test: %s", err)
		}

		log.Printf("Tests Succeeded!\n\n")

		err = lang.Build(workDir, meta, buildSkipTargets, local)
		if err != nil {
			log.Fatalf("build failed: %s", err)
		}

		err = gm.HandleArtifacts(meta, workDir, cwd, true, false, true, buildSkipTargets, local)
		if err != nil {
			log.Fatalf("signing failed: %s", err)
		}

		err = gm.HandleExtras(meta, workDir, cwd, true, false, true, local)
		if err != nil {
			log.Fatalf("Extra artifact processing failed: %s", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(signCmd)
}
