package services

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/goodsign/monday"
	"gopkg.in/yaml.v3"
)

type (
	Translator struct {
		translationMap map[string]interface{}
		TimeLanguage   monday.Locale
	}
)

var timeLocales = map[string]monday.Locale{
	"en": monday.LocaleEnUS,
	"es": monday.LocaleEsES,
}

// NewTranslator will load a translation file and return a new template renderer.
func NewTranslator(path string, lang string) (*Translator, error) {
	translation, err := os.ReadFile(path + "/translation." + lang + ".yaml")
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
		TimeLanguage:   getTimeLocale(lang),
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

// TransDate returns a translated date with the specified format.
func (r *Translator) TransDate(d time.Time, f string) string {
	return monday.Format(d, f, r.TimeLanguage)
}

// getTimeLocale returns the monday.Locale corresponding to lang.
// If lang is not supported by monday, it returns the monday.LocaleEnUS locale.
func getTimeLocale(lang string) monday.Locale {
	if locale, ok := timeLocales[lang]; ok {
		return locale
	}

	log.Printf("Monday: locale %q not supported, using default (en_US)", lang)
	return monday.LocaleEnUS
}
