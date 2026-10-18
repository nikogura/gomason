// Copyright © 2026 Nik Ogura <nik.ogura@gmail.com>
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
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmdStructure(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "gomason", rootCmd.Use, "root command should be invoked as gomason")
	assert.Contains(t, rootCmd.Short, "Go binaries", "short description should say what it builds")
	assert.NotEmpty(t, rootCmd.Long, "long description should be set")
	assert.NotNil(t, rootCmd.PersistentPreRun, "verbose handling hangs off PersistentPreRun")
}

func TestRootCmdPersistentFlags(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		flag      string
		shorthand string
		defValue  string
	}{
		{name: "verbose", flag: "verbose", shorthand: "v", defValue: "false"},
		{name: "dryrun", flag: "dryrun", shorthand: "d", defValue: "false"},
		{name: "branch", flag: "branch", shorthand: "b", defValue: ""},
		{name: "workdir", flag: "workdir", shorthand: "w", defValue: ""},
		{name: "local", flag: "local", shorthand: "l", defValue: "false"},
		{name: "skip build targets", flag: "skip-build-targets", shorthand: "", defValue: ""},
		{name: "test timeout", flag: "test-timeout", shorthand: "", defValue: ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			f := rootCmd.PersistentFlags().Lookup(c.flag)
			require.NotNil(t, f, "%s flag should exist", c.flag)
			assert.Equal(t, c.shorthand, f.Shorthand, "%s shorthand", c.flag)
			assert.Equal(t, c.defValue, f.DefValue, "%s default", c.flag)
		})
	}
}

func TestSubcommandsRegistered(t *testing.T) {
	t.Parallel()

	registered := make(map[string]*cobra.Command)
	for _, c := range rootCmd.Commands() {
		registered[c.Name()] = c
	}

	for _, name := range []string{"build", "test", "sign", "publish", "version"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cmd, ok := registered[name]
			require.True(t, ok, "%s should be registered on the root command", name)
			assert.NotEmpty(t, cmd.Short, "%s should have a short description", name)
			assert.NotNil(t, cmd.Run, "%s should have a Run function", name)
		})
	}
}

func TestCommandFlags(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		cmd       *cobra.Command
		flag      string
		shorthand string
		defValue  string
	}{
		{name: "build skiptests", cmd: buildCmd, flag: "skiptests", shorthand: "s", defValue: "false"},
		{name: "publish skiptests", cmd: publishCmd, flag: "skiptests", shorthand: "s", defValue: "false"},
		{name: "publish skipbuild", cmd: publishCmd, flag: "skipbuild", shorthand: "", defValue: "false"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			f := c.cmd.Flags().Lookup(c.flag)
			require.NotNil(t, f, "%s should exist on %s", c.flag, c.cmd.Name())
			assert.Equal(t, c.shorthand, f.Shorthand, "%s shorthand", c.flag)
			assert.Equal(t, c.defValue, f.DefValue, "%s default", c.flag)
		})
	}
}

// TestRootCmdHelp drives the real command, so it mutates rootCmd's output and
// args.  It deliberately does not run in parallel with the flag tests above.
func TestRootCmdHelp(t *testing.T) {
	buf := new(bytes.Buffer)

	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--help"})

	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetArgs(nil)
	})

	err := rootCmd.Execute()
	require.NoError(t, err, "help should not error")

	output := buf.String()

	assert.Contains(t, output, "gomason", "help should name the binary")

	for _, sub := range []string{"build", "test", "sign", "publish", "version"} {
		assert.Contains(t, output, sub, "help should list the %s subcommand", sub)
	}

	for _, flag := range []string{"--verbose", "--branch", "--workdir", "--local"} {
		assert.Contains(t, output, flag, "help should list %s", flag)
	}
}

func TestVersionCmdStructure(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "version", versionCmd.Use, "version command should be invoked as version")
	assert.Contains(t, versionCmd.Short, "Version", "short description should mention version")
	assert.NotNil(t, versionCmd.Run, "version command should have a Run function")
}
