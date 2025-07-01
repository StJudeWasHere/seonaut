package services

import (
	"fmt"
	"log"
	"os"
	"strings"

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

// Trans returns a translated string with numbered parameters replaced.
func (r *Translator) Trans(s string, args ...interface{}) string {
	t, ok := r.translationMap[s]
	if !ok {
		log.Printf("trans: %s translation not found\n", s)
		return s
	}

	result := fmt.Sprintf("%v", t)
	for i, arg := range args {
		placeholder := fmt.Sprintf("%%%d%%", i+1)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", arg))
	}
	return result
}
