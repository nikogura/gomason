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
	"fmt"
	"os"

	"github.com/nikogura/gomason/v2/pkg/gomason"
)

// workDirPrefix is the prefix given to gomason's ephemeral work directories.
const workDirPrefix = "gomason"

// createEphemeralWorkDir creates the temporary root directory a gomason
// command builds in.  The caller owns it, and must hand it to cleanupWorkDir
// when finished.
func createEphemeralWorkDir() (rootWorkDir string, err error) {
	rootWorkDir, err = os.MkdirTemp("", workDirPrefix)
	if err != nil {
		err = fmt.Errorf("failed to create temp dir: %w", err)

		return rootWorkDir, err
	}

	return rootWorkDir, err
}

// cleanupWorkDir removes an ephemeral root work directory.
//
// An empty path is a no-op.  That is what a local build passes, since a local
// build works out of the current working directory and has nothing of its own
// to remove -- removing the user's working directory would be catastrophic.
//
// Cleanup is called on the success and failure paths alike, and deliberately
// not deferred: gomason's commands exit the process on error, and a deferred
// cleanup does not run through os.Exit.
func cleanupWorkDir(rootWorkDir string) (err error) {
	if rootWorkDir == "" {
		return err
	}

	err = os.RemoveAll(rootWorkDir)
	if err != nil {
		err = fmt.Errorf("failed to clean up work dir %s: %w", rootWorkDir, err)

		return err
	}

	return err
}

// prepareWorkspace creates a clean, ephemeral GOPATH and checks the project
// out into it.  It returns the root work dir even when it fails partway, so
// the caller can always clean up what was created.
func prepareWorkspace(lang gomason.Language, meta gomason.Metadata) (rootWorkDir string, workDir string, err error) {
	rootWorkDir, err = createEphemeralWorkDir()
	if err != nil {
		return rootWorkDir, workDir, err
	}

	workDir, err = lang.CreateWorkDir(rootWorkDir)
	if err != nil {
		err = fmt.Errorf("failed to create ephemeral workDir: %w", err)

		return rootWorkDir, workDir, err
	}

	err = lang.Checkout(workDir, meta, branch)
	if err != nil {
		err = fmt.Errorf("failed to checkout package %s at branch %s: %w", meta.Package, branch, err)

		return rootWorkDir, workDir, err
	}

	return rootWorkDir, workDir, err
}
