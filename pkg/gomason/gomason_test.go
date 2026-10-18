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
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/nikogura/gomason/v2/pkg/gomason/testfixtures"
	"github.com/phayes/freeport"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

//nolint:gochecknoglobals // test fixture shared state
var TestTmpDir string

//nolint:gochecknoglobals // test fixture shared state
var servicePort int

func TestMain(m *testing.M) {
	setUp()

	code := m.Run()

	tearDown()

	os.Exit(code)
}

func setUp() {
	logrus.SetLevel(logrus.DebugLevel)

	dir, err := os.MkdirTemp("", "gomason")
	if err != nil {
		log.Fatal("Error creating temp dir\n")
	}

	TestTmpDir = dir

	fmt.Printf("Setting up global temporary work dir %s\n", TestTmpDir)

	freePort, err := freeport.GetFreePort()
	if err != nil {
		fmt.Printf("Error getting a free port: %s\n", err)
		os.Exit(1)
	}

	servicePort = freePort

	tr := testfixtures.TestRepo{}

	go tr.Run(servicePort)

}

func tearDown() {
	_, statErr := os.Stat(TestTmpDir)
	if !os.IsNotExist(statErr) {
		_ = os.Remove(TestTmpDir)
	}
}

func TestGetUserConfig(t *testing.T) {
	homeDir := fmt.Sprintf("%s/user", TestTmpDir)

	err := os.Mkdir(homeDir, 0755)
	if err != nil {
		t.Errorf("Failed creating %s", homeDir)
	}

	userFile := fmt.Sprintf("%s/.gomason", homeDir)

	err = os.WriteFile(userFile, []byte(testfixtures.UserConfig), 0644)
	if err != nil {
		t.Errorf("error writing %s: %s", userFile, err)
	}

	actual, err := GetUserConfig(homeDir)
	if err != nil {
		t.Errorf("Error getting user config: %s", err)
	}

	expected := UserConfig{
		User: UserInfo{
			Email:        "nik.ogura@gmail.com",
			Username:     "nikogura",
			UsernameFunc: "echo 'foo bar baz'",
			Password:     "changeit",
			PasswordFunc: "echo 'seecret!'",
		},
		Signing: UserSignInfo{
			Program: "gpg",
		},
	}

	assert.Equal(t, expected, actual, "loaded config meets expectations")
}

func TestGetUserConfigWithBearerToken(t *testing.T) {
	homeDir := fmt.Sprintf("%s/user-bearer", TestTmpDir)

	err := os.Mkdir(homeDir, 0755)
	if err != nil {
		t.Errorf("Failed creating %s", homeDir)
	}

	userFile := fmt.Sprintf("%s/.gomason", homeDir)

	err = os.WriteFile(userFile, []byte(testfixtures.UserConfigWithBearerToken), 0644)
	if err != nil {
		t.Errorf("error writing %s: %s", userFile, err)
	}

	actual, err := GetUserConfig(homeDir)
	if err != nil {
		t.Errorf("Error getting user config: %s", err)
	}

	expected := UserConfig{
		User: UserInfo{
			Email:       "nik.ogura@gmail.com",
			Username:    "nikogura",
			Password:    "changeit",
			BearerToken: "static-test-token-12345",
		},
		Signing: UserSignInfo{
			Program: "gpg",
		},
	}

	assert.Equal(t, expected, actual, "loaded config with bearer token meets expectations")
}

func TestGetUserConfigWithBearerTokenFunc(t *testing.T) {
	homeDir := fmt.Sprintf("%s/user-bearer-func", TestTmpDir)

	err := os.Mkdir(homeDir, 0755)
	if err != nil {
		t.Errorf("Failed creating %s", homeDir)
	}

	userFile := fmt.Sprintf("%s/.gomason", homeDir)

	err = os.WriteFile(userFile, []byte(testfixtures.UserConfigWithBearerTokenFunc), 0644)
	if err != nil {
		t.Errorf("error writing %s: %s", userFile, err)
	}

	actual, err := GetUserConfig(homeDir)
	if err != nil {
		t.Errorf("Error getting user config: %s", err)
	}

	expected := UserConfig{
		User: UserInfo{
			Email:           "nik.ogura@gmail.com",
			BearerTokenFunc: "echo 'dynamic-token-from-func'",
		},
		Signing: UserSignInfo{
			Program: "gpg",
		},
	}

	assert.Equal(t, expected, actual, "loaded config with bearer token func meets expectations")
}
