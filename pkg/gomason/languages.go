package gomason

import (
	"fmt"
)

const (
	LanguageGolang = "golang"
)

// Language is a generic interface for doing what gomason does - abstracting build, test, signing, and publishing of binaries
type Language interface {
	CreateWorkDir(string) (string, error)
	Checkout(string, Metadata, string) error
	Prep(string, Metadata) error
	Test(string, string) error
	Build(string, Metadata, string) error
}

// NoLanguage essentially an abstract class for the Language interface
type NoLanguage struct{}

// CreateWorkDir Stub for the CreateWorkDir action
func (NoLanguage) CreateWorkDir(string) (string, error) {
	return "", nil
}

// Checkout Stub for the Checkout action
func (NoLanguage) Checkout(string, Metadata, string) error {
	return nil
}

// Prep stub for the Prep action
func (NoLanguage) Prep(string, Metadata) error {
	return nil
}

// Test Stub for the Test action
func (NoLanguage) Test(string, string) error {
	return nil
}

// Build Stub for the Build Action
func (NoLanguage) Build(string, Metadata, string) error {
	return nil
}

var languagesMap map[string]Language = map[string]Language{}

// GetByName Gets a specific Language interface by name.
func GetByName(lang string) (Language, error) {
	l, ok := languagesMap[lang]
	if !ok {
		return NoLanguage{}, fmt.Errorf("Unsupported language: %s", lang)
	}
	return l, nil
}
