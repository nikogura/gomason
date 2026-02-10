package gomason

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	// LanguageGolang is a canonical string representation of the golang language.
	LanguageGolang = "golang"
)

// Language is a generic interface for doing what gomason does - abstracting build, test, signing, and publishing of binaries.
type Language interface {
	CreateWorkDir(string) (string, error)
	Checkout(workdir string, meta Metadata, branch string) error
	Prep(workdir string, meta Metadata, local bool) error
	Test(workdir string, module string, timeout string, local bool) error
	Build(workdir string, meta Metadata, skipTargets string, local bool) error
}

// NoLanguage is essentially an abstract class for the Language interface.
type NoLanguage struct{}

// CreateWorkDir is a stub for the CreateWorkDir action.
func (NoLanguage) CreateWorkDir(string) (workdir string, err error) {
	return workdir, err
}

// Checkout is a stub for the Checkout action.
func (NoLanguage) Checkout(workdir string, meta Metadata, branch string) (err error) {
	return err
}

// Prep is a stub for the Prep action.
func (NoLanguage) Prep(workdir string, meta Metadata, local bool) (err error) {
	return err
}

// Test is a stub for the Test action.
func (NoLanguage) Test(workdir string, module string, timeout string, localTest bool) (err error) {
	return err
}

// Build is a stub for the Build action.
func (NoLanguage) Build(workdor string, meta Metadata, skipTargets string, localBuild bool) (err error) {
	return err
}

//nolint:gochecknoglobals // language registry pattern
var languagesMap map[string]Language = map[string]Language{}

// GetByName Gets a specific Language interface by name.
func GetByName(lang string) (language Language, err error) {
	var ok bool

	language, ok = languagesMap[lang]
	if !ok {
		language = NoLanguage{}
		err = errors.New(fmt.Sprintf("Unsupported language: %s", lang))

		return language, err
	}

	return language, err
}
