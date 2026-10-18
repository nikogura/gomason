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
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nikogura/gomason/v2/pkg/gomason"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEphemeralWorkDir(t *testing.T) {
	t.Parallel()

	rootWorkDir, err := createEphemeralWorkDir()
	require.NoError(t, err, "creating an ephemeral work dir should succeed")
	require.DirExists(t, rootWorkDir, "the created work dir should exist")

	assert.True(t, strings.HasPrefix(filepath.Base(rootWorkDir), workDirPrefix),
		"work dir %q should carry the %q prefix", rootWorkDir, workDirPrefix)

	cleanupErr := cleanupWorkDir(rootWorkDir)
	require.NoError(t, cleanupErr, "cleaning up the work dir should succeed")
}

func TestCleanupWorkDirRemovesPopulatedTree(t *testing.T) {
	t.Parallel()

	rootWorkDir, err := createEphemeralWorkDir()
	require.NoError(t, err)

	// Cleanup has to cope with a full checkout, not just an empty directory.
	nested := filepath.Join(rootWorkDir, "go", "src", "github.com", "nikogura", "testproject")
	require.NoError(t, os.MkdirAll(nested, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(nested, "main.go"), []byte("package main\n"), 0644))

	err = cleanupWorkDir(rootWorkDir)
	require.NoError(t, err, "cleanup should remove a populated tree")
	assert.NoDirExists(t, rootWorkDir, "work dir should be gone after cleanup")
}

func TestCleanupWorkDirEmptyPathLeavesWorkingDirAlone(t *testing.T) {
	// A local build passes an empty path, because it works out of the user's
	// current directory and has nothing of its own to remove.  Removing the
	// working directory instead would be catastrophic, so pin the behavior.
	dir := t.TempDir()
	t.Chdir(dir)

	sentinel := filepath.Join(dir, "please-do-not-delete-me")
	require.NoError(t, os.WriteFile(sentinel, []byte("important"), 0644))

	err := cleanupWorkDir("")
	require.NoError(t, err, "an empty path should be a no-op, not an error")

	assert.FileExists(t, sentinel, "cleanup must not touch the working directory")
	assert.DirExists(t, dir, "cleanup must not remove the working directory")
}

func TestCleanupWorkDirMissingPathIsNotAnError(t *testing.T) {
	t.Parallel()

	// Cleanup runs on the failure paths too, where the tree may already be
	// partially gone.  Removing something that isn't there is fine.
	err := cleanupWorkDir(filepath.Join(t.TempDir(), "never-created"))
	assert.NoError(t, err, "removing an absent path should not be an error")
}

// failingLanguage is a gomason.Language whose checkout fails, after a work
// dir has already been created.  It stands in for the real-world case the
// leak regression test cares about: a checkout that dies partway, leaving a
// populated tree behind.
type failingLanguage struct{}

func (failingLanguage) CreateWorkDir(workDir string) (gopath string, err error) {
	gopath = filepath.Join(workDir, "go")
	err = os.MkdirAll(gopath, 0755)

	return gopath, err
}

func (failingLanguage) Checkout(workdir string, meta gomason.Metadata, branch string) (err error) {
	err = errors.New("checkout blew up")

	return err
}

func (failingLanguage) Prep(workdir string, meta gomason.Metadata, local bool) (err error) {
	return err
}

func (failingLanguage) Test(workdir string, module string, timeout string, local bool) (err error) {
	return err
}

func (failingLanguage) Build(workdir string, meta gomason.Metadata, skipTargets string, local bool) (err error) {
	return err
}

// TestPrepareWorkspaceReportsWorkDirOnFailure is the regression test for the
// leaked work dir.  prepareWorkspace must hand back the ephemeral dir it
// created even when a later step fails, so the caller can always clean it up.
// Before this, cleanup was deferred and log.Fatalf walked straight past it,
// stranding a full GOPATH in the temp dir on every failed run.
func TestPrepareWorkspaceReportsWorkDirOnFailure(t *testing.T) {
	t.Parallel()

	rootWorkDir, workDir, err := prepareWorkspace(failingLanguage{}, gomason.Metadata{
		Package: "github.com/nikogura/does-not-exist",
	})

	require.Error(t, err, "the stub language fails its checkout")
	require.ErrorContains(t, err, "failed to checkout", "the failure should come from the checkout step")

	require.NotEmpty(t, rootWorkDir, "a failed prepare must still report its work dir")
	require.DirExists(t, rootWorkDir, "the work dir exists, and is leaked without cleanup")
	require.DirExists(t, workDir, "the checkout tree was created before the failure")

	cleanupErr := cleanupWorkDir(rootWorkDir)
	require.NoError(t, cleanupErr)
	assert.NoDirExists(t, rootWorkDir, "the reported work dir should be removable")
}

// TestRunTestCreatesNothingBeforeReadingMetadata pins the ordering: a run that
// dies on a missing metadata.json has not created a work dir yet, so it
// reports an empty path and there is nothing to clean up.
func TestRunTestCreatesNothingBeforeReadingMetadata(t *testing.T) {
	// An empty directory has no metadata.json.
	t.Chdir(t.TempDir())

	rootWorkDir, err := runTest()

	require.Error(t, err, "runTest should fail without a metadata file")
	require.ErrorContains(t, err, "failed to read metadata")
	assert.Empty(t, rootWorkDir, "nothing was created, so nothing is reported")
}
