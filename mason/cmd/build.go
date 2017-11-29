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
	"github.com/nikogura/gomason/mason"
	"github.com/spf13/cobra"
	"log"
)

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
		_, err := mason.WholeShebang(workdir, branch, true, false, false, verbose)
		if err != nil {
			log.Fatalf("Error running test and build: %s\n", err)
		}

	},
}

func init() {
	RootCmd.AddCommand(buildCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// buildCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
