package gomason

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/phayes/freeport"
)

var TestTmpDir string
var servicePort int

func TestMain(m *testing.M) {
	setUp()

	code := m.Run()

	tearDown()

	os.Exit(code)
}

func setUp() {
	dir, err := ioutil.TempDir("", "gomason")
	if err != nil {
		log.Fatal("Error creating temp dir\n")
	}

	TestTmpDir = dir

	log.Printf("Setting up temporary work dir %s", TestTmpDir)

	freePort, err := freeport.GetFreePort()
	if err != nil {
		log.Printf("Error getting a free port: %s", err)
		os.Exit(1)
	}

	servicePort = freePort

	tr := TestRepo{}

	go tr.Run(servicePort)

}

func tearDown() {
	if _, err := os.Stat(TestTmpDir); !os.IsNotExist(err) {
		os.Remove(TestTmpDir)
	}
}

func TestNoLanguage(t *testing.T) {
	nl := NoLanguage{}

	_, err := nl.CreateWorkDir("")
	assert.True(t, err == nil, "Create work dir returned an error")

	err = nl.Checkout("", Metadata{}, "", true)
	assert.True(t, err == nil, "Checkout returned an error")

	err = nl.Prep("", Metadata{}, true)
	assert.True(t, err == nil, "Prep returned an error")

	err = nl.Test("", "", true)
	assert.True(t, err == nil, "Test returned an error")

	err = nl.Build("", Metadata{}, "", true)
	assert.True(t, err == nil, "Build returned an error")

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
			"Unsupported language: foo",
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
