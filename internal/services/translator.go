package services

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type (
	TranslatorConfig struct {
		TranslationsFile string
	}

	Translator struct {
		translationMap map[string]interface{}
		config         *TranslatorConfig
	}
)

// NewTranslator will load a translation file and return a new template renderer.
func NewTranslator(config *TranslatorConfig) (*Translator, error) {
	translation, err := os.ReadFile(config.TranslationsFile)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	err = yaml.Unmarshal(translation, &m)
	if err != nil {
		return nil, err
	}

	return &Translator{
		translationMap: m,
		config:         config,
	}, nil
}

// Returns a translated string from the translations map.
// The original string is returned if a translation is not found in the map.
func (r *Translator) Trans(s string) string {
	t, ok := r.translationMap[s]
	if !ok {
		log.Printf("trans: %s translation not found\n", s)
		return s
	}

	return fmt.Sprintf("%v", t)
}
