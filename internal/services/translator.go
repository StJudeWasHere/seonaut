package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goodsign/monday"
	"gopkg.in/yaml.v3"
)

type (
	Translator struct {
		translationMap    map[string]map[string]interface{}
		defaultLang       string
		defaultTimeLocale monday.Locale
	}
)

// Language conversion table for the monday library
var timeLocales = map[string]monday.Locale{
	"en": monday.LocaleEnUS, // English (United States)
	"da": monday.LocaleDaDK, // Danish (Denmark)
	"nl": monday.LocaleNlNL, // Dutch (Netherlands)
	"fi": monday.LocaleFiFI, // Finnish (Finland)
	"fr": monday.LocaleFrFR, // French (France)
	"de": monday.LocaleDeDE, // German (Germany)
	"hu": monday.LocaleHuHU, // Hungarian (Hungary)
	"it": monday.LocaleItIT, // Italian (Italy)
	"nn": monday.LocaleNnNO, // Norwegian Nynorsk (Norway)
	"nb": monday.LocaleNbNO, // Norwegian Bokmål (Norway)
	"pl": monday.LocalePlPL, // Polish (Poland)
	"pt": monday.LocalePtPT, // Portuguese (Portugal)
	"ro": monday.LocaleRoRO, // Romanian (Romania)
	"ru": monday.LocaleRuRU, // Russian (Russia)
	"es": monday.LocaleEsES, // Spanish (Spain)
	"ca": monday.LocaleCaES, // Catalan (Spain)
	"sv": monday.LocaleSvSE, // Swedish (Sweden)
	"tr": monday.LocaleTrTR, // Turkish (Turkey)
	"uk": monday.LocaleUkUA, // Ukrainian (Ukraine)
	"bg": monday.LocaleBgBG, // Bulgarian (Bulgaria)
	"zh": monday.LocaleZhCN, // Chinese (Mainland)
	"ko": monday.LocaleKoKR, // Korean (Korea)
	"ja": monday.LocaleJaJP, // Japanese (Japan)
	"el": monday.LocaleElGR, // Greek (Greece)
	"id": monday.LocaleIdID, // Indonesian (Indonesia)
	"cs": monday.LocaleCsCZ, // Czech (Czech Republic)
	"sl": monday.LocaleSlSI, // Slovenian (Slovenia)
	"lt": monday.LocaleLtLT, // Lithuanian (Lithuania)
	"et": monday.LocaleEtEE, // Estonian (Estonia)
	"hr": monday.LocaleHrHR, // Croatian (Croatia)
	"lv": monday.LocaleLvLV, // Latvian (Latvia)
	"sk": monday.LocaleSkSK, // Slovak (Slovakia)
	"th": monday.LocaleThTH, // Thai (Thailand)
	"uz": monday.LocaleUzUZ, // Uzbek (Uzbekistan)
	"kk": monday.LocaleKkKZ, // Kazakh (Kazakhstan)
}

// NewTranslator will load all translation files in the path and set the default language specified
// in the defaultLang parameter.
// If the default language does not exist it returns an error, otherwise it will return a new Translator.
// Translation files must follow follow this naming convention: translation.(lang_code).yaml replacing (lang_code)
// with a two letter language code.
func NewTranslator(path string, defaultLang string) (*Translator, error) {
	t := &Translator{
		translationMap:    make(map[string]map[string]interface{}),
		defaultLang:       defaultLang,
		defaultTimeLocale: getTimeLocale(defaultLang),
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("error reading translations directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasPrefix(file.Name(), "translation.") || !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		parts := strings.Split(file.Name(), ".")
		if len(parts) != 3 {
			continue
		}
		lang := parts[1]

		translation, err := os.ReadFile(filepath.Join(path, file.Name()))
		if err != nil {
			log.Printf("warning: could not read translation file %s: %v\n", file.Name(), err)
			continue
		}

		m := make(map[string]interface{})
		err = yaml.Unmarshal(translation, &m)
		if err != nil {
			return nil, err
		}
		t.translationMap[lang] = m
	}

	if _, ok := t.translationMap[defaultLang]; !ok {
		return nil, fmt.Errorf("default language %q not found in translations", defaultLang)
	}

	return t, nil
}

// Trans returns a translated string with numbered parameters replaced.
func (r *Translator) Trans(lang, s string, args ...interface{}) string {
	translationMap, ok := r.translationMap[lang]
	if !ok {
		log.Printf("trans: locale %q not found, using default (%s)\n", lang, r.defaultLang)
		translationMap = r.translationMap[r.defaultLang]
	}

	t, ok := translationMap[s]
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
func (r *Translator) TransDate(lang string, d time.Time, f string) string {
	return monday.Format(d, f, getTimeLocale(lang))
}

// LangIsSupported checks if a language is supported by checking if it is loaded in the translationMap.
// It returns true if the language is loaded and false otherwise.
func (r *Translator) LangIsSupported(lang string) bool {
	_, ok := r.translationMap[lang]

	return ok
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
