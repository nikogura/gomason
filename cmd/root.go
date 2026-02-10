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
	"github.com/sirupsen/logrus"
	"os"

	"github.com/spf13/cobra"
)

// Verbose sets verbose output.
//
//nolint:gochecknoglobals // Cobra boilerplate
var Verbose bool

//nolint:gochecknoglobals // Cobra boilerplate
var branch string

//nolint:gochecknoglobals // Cobra boilerplate
var dryrun bool

//nolint:gochecknoglobals // Cobra boilerplate
var workdir string

//nolint:gochecknoglobals // Cobra boilerplate
var buildSkipTargets string

//nolint:gochecknoglobals // Cobra boilerplate
var testTimeout string

//nolint:gochecknoglobals // Cobra boilerplate
var local bool

// rootCmd represents the base command when called without any subcommands.
//
//nolint:gochecknoglobals // Cobra boilerplate
var rootCmd = &cobra.Command{
	Use:   "gomason",
	Short: "Tool for building Go binaries in a clean GOPATH.",
	Long: `
Tool for building Go binaries in a clean GOPATH.

`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if Verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}
	},
}

// Execute runs the root cobra command.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

//nolint:gochecknoinits // Cobra boilerplate
func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVarP(&dryrun, "dryrun", "d", false, "Dry Run (Only applies to publish.")
	rootCmd.PersistentFlags().StringVarP(&branch, "branch", "b", "", "Branch to operate upon")
	rootCmd.PersistentFlags().StringVarP(&workdir, "workdir", "w", "", "Workdir.  If omitted, a temp dir will be created and subsequently cleaned up.")
	rootCmd.PersistentFlags().StringVarP(&buildSkipTargets, "skip-build-targets", "", "", fmt.Sprintf("Comma separated list of build targets from %s to skip.", gomason.MetadataFilename))
	rootCmd.PersistentFlags().StringVarP(&testTimeout, "test-timeout", "", "", "timeout for tests to complete (must be valid time input for language)")

	rootCmd.PersistentFlags().BoolVarP(&local, "local", "l", false, "Do all work out of current working directory, with whatever is checked out.")
	//rootCmd.PersistentFlags().StringVarP(&pubSkipTargets, fmt.Sprintf("skip-publish-targets", "", "", "Comma separated list of publish targets from %s to skip.", gomason.MetadataFilename))
}
