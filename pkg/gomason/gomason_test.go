package gomason

import (
	"fmt"
	"github.com/phayes/freeport"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
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

	tr := TestRepo{}

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

	err = os.WriteFile(userFile, []byte(testUserConfig()), 0644)
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

	err = os.WriteFile(userFile, []byte(testUserConfigWithBearerToken()), 0644)
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

	err = os.WriteFile(userFile, []byte(testUserConfigWithBearerTokenFunc()), 0644)
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
