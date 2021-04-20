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
	"github.com/nikogura/gomason/pkg/gomason"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test your code in a clean environment.",
	Long: `
Test your code in a clean environment.

You know how it goes, you write stuff.  You even test it.  You commit it, you push it,
and then you get nagging and embarrassing issues logged against your otherwise wonderful project because you forgot to
list some code dependency or other.

Gomason will help protect you from such infamy by building your code in a clean environment locally and letting you know the results.

Sure, you could do the same thing with a CI or CD system.  But sometimes that's not an option.

Sometimes you need the benefits of a full system here.  Now.  Right at your fingertips.  You're welcome.
`,
	Run: func(cmd *cobra.Command, args []string) {
		_, err := gomason.NewGomason()
		if err != nil {
			log.Fatalf("error creating gomason object")
		}

		rootWorkDir, err := ioutil.TempDir("", "gomason")
		if err != nil {
			log.Fatalf("Failed to create temp dir: %s", err)
		}

		log.Printf("[DEBUG] Created temp dir %s", rootWorkDir)

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

		err = lang.Test(workDir, meta.Package, testTimeout)
		if err != nil {
			log.Fatalf("error running go test: %s", err)
		}

		fmt.Print("Success!\n\n")
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
