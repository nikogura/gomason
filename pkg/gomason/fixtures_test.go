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

package gomason

import (
	"github.com/nikogura/gomason/v2/pkg/gomason/testfixtures"
)

// testMetadataObj returns a Metadata object suitable for testing.  It lives
// here rather than in the testfixtures package because it returns a gomason
// type, and testfixtures deliberately does not import gomason.  Keep it in
// step with testfixtures.MetadataJSON.
func testMetadataObj() (metadata Metadata) {
	target := PublishTarget{
		Source:      testfixtures.PublishSource,
		Destination: testfixtures.PublishDestination,
		Signature:   true,
		Checksums:   true,
	}

	metadata = Metadata{
		Package:     testfixtures.ModuleName,
		Version:     testfixtures.Version,
		Description: testfixtures.Description,
		BuildInfo: BuildInfo{
			PrepCommands: []string{
				testfixtures.PrepCommand,
			},
			Targets: []BuildTarget{
				{
					Name: testfixtures.BuildTargetName,
					Flags: map[string]string{
						testfixtures.BuildFlagName: testfixtures.BuildFlagValue,
					},
				},
			},
		},
		SignInfo: SignInfo{
			Program: testfixtures.SigningProgram,
			Email:   testfixtures.SigningEmail,
		},
		PublishInfo: PublishInfo{
			Targets: []PublishTarget{target},
			TargetsMap: map[string]PublishTarget{
				testfixtures.PublishSource: target,
			},
		},
	}

	return metadata
}

// testMetadataFileName returns the metadata file name gomason looks for.
func testMetadataFileName() (name string) {
	name = MetadataFilename

	return name
}
