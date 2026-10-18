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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetadataGetLanguage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		meta     Metadata
		expected string
	}{
		{
			name:     "explicit golang",
			meta:     Metadata{Language: LanguageGolang},
			expected: LanguageGolang,
		},
		{
			name:     "non-default language is preserved",
			meta:     Metadata{Language: "python"},
			expected: "python",
		},
		{
			name:     "empty language defaults to golang",
			meta:     Metadata{Language: ""},
			expected: LanguageGolang,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := c.meta.GetLanguage()

			assert.Equal(t, c.expected, actual, "returned language meets expectations")
		})
	}
}
