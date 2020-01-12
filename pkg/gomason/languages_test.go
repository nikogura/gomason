package gomason

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoLanguage(t *testing.T) {
	nl := NoLanguage{}

	_, err := nl.CreateWorkDir("")
	assert.True(t, err == nil, "Create work dir returned an error")

	err = nl.Checkout("", Metadata{}, "")
	assert.True(t, err == nil, "Checkout returned an error")

	err = nl.Prep("", Metadata{})
	assert.True(t, err == nil, "Prep returned an error")

	err = nl.Test("", "")
	assert.True(t, err == nil, "Test returned an error")

	err = nl.Build("", Metadata{})
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
