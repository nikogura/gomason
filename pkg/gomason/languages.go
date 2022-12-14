package gomason

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	// LanguageGolang a canonical string representation of the golang language
	LanguageGolang = "golang"
)

// Language is a generic interface for doing what gomason does - abstracting build, test, signing, and publishing of binaries
type Language interface {
	CreateWorkDir(string) (string, error)
	Checkout(workdir string, meta Metadata, branch string) error
	Prep(workdir string, meta Metadata) error
	Test(workdir string, module string, timeout string) error
	Build(workdir string, meta Metadata, skipTargets string) error
}

// NoLanguage essentially an abstract class for the Language interface
type NoLanguage struct{}

// CreateWorkDir Stub for the CreateWorkDir action
func (NoLanguage) CreateWorkDir(string) (workdir string, err error) {
	return workdir, err
}

// Checkout Stub for the Checkout action
func (NoLanguage) Checkout(workdir string, meta Metadata, branch string) error {
	return nil
}

// Prep stub for the Prep action
func (NoLanguage) Prep(workdir string, meta Metadata) error {
	return nil
}

// Test Stub for the Test action
func (NoLanguage) Test(workdir string, module string, timeout string) error {
	return nil
}

// Build Stub for the Build Action
func (NoLanguage) Build(workdor string, meta Metadata, skipTargets string) error {
	return nil
}

var languagesMap map[string]Language = map[string]Language{}

// GetByName Gets a specific Language interface by name.
func GetByName(lang string) (Language, error) {
	l, ok := languagesMap[lang]
	if !ok {
		return NoLanguage{}, errors.New(fmt.Sprintf("Unsupported language: %s", lang))
	}
	return l, nil
}
