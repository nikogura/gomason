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

package gomason

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoLanguage(t *testing.T) {
	nl := NoLanguage{}

	_, err := nl.CreateWorkDir("")
	require.NoError(t, err, "Create work dir returned an error")

	err = nl.Checkout("", Metadata{}, "")
	require.NoError(t, err, "Checkout returned an error")

	err = nl.Prep("", Metadata{}, false)
	require.NoError(t, err, "Prep returned an error")

	err = nl.Test("", "", "", false)
	require.NoError(t, err, "Test returned an error")

	err = nl.Build("", Metadata{}, "", false)
	require.NoError(t, err, "Build returned an error")

}

func TestGetByName(t *testing.T) {
	var inputs = []struct {
		name        string
		input       string
		output      interface{}
		errorstring string
	}{
		{
			"unsupported",
			"foo",
			NoLanguage{},
			"unsupported language: foo",
		},
		{
			"golang",
			"golang",
			Golang{},
			"",
		},
	}

	for _, tc := range inputs {
		t.Run(tc.name, func(t *testing.T) {
			iface, err := GetByName(tc.input)
			assert.True(t, reflect.DeepEqual(iface, tc.output), "Interface mismatch at %s", tc.name)
			if err != nil {
				assert.Equal(t, err.Error(), tc.errorstring, "Error does not meet expectations.")
			}
		})
	}

}
