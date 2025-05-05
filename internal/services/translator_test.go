package services_test

import (
	"testing"

	"github.com/stjudewashere/seonaut/internal/services"
)

// Test the translator's Trans method
func TestTrans(t *testing.T) {
	translator, err := services.NewTranslator(&services.TranslatorConfig{
		TranslationsFile: "./testdata/translations.test.yaml",
	})
	if err != nil {
		t.Fatalf("%v", err)
	}

	got := translator.Trans("TEST")
	want := "test translation"
	if got != want {
		t.Errorf("renderer %s != %s", got, want)
	}
}
