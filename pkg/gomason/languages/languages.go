package languages

import (
	"fmt"

	"github.com/nikogura/gomason/pkg/gomason"
)

type Language interface {
	CreateWorkDir(string) (string, error)
	Checkout(string, gomason.Metadata, string, bool) error
	Prep(string, gomason.Metadata, bool) error
	Test(string, string, bool) error
	Build(string, gomason.Metadata, string, bool) error
}

type NoLanguage struct{}

func (NoLanguage) CreateWorkDir(string) (string, error) {
	return "", nil
}

func (NoLanguage) Checkout(string, gomason.Metadata, string, bool) error {
	return nil
}

func (NoLanguage) Prep(string, gomason.Metadata, bool) error {
	return nil
}

func (NoLanguage) Test(string, string, bool) error {
	return nil
}

func (NoLanguage) Build(string, gomason.Metadata, string, bool) error {
	return nil
}

var languagesMap map[string]Language = map[string]Language{}

func GetByName(lang string) (Language, error) {
	l, ok := languagesMap[lang]
	if !ok {
		return NoLanguage{}, fmt.Errorf("Unsupported language: %s", lang)
	}
	return l, nil
}
