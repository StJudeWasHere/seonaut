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

// Test the translator's Trans method with parameters
func TestTransParameters(t *testing.T) {
	translator, err := services.NewTranslator(&services.TranslatorConfig{
		TranslationsFile: "./testdata/translations.test.yaml",
	})
	if err != nil {
		t.Fatalf("%v", err)
	}

	got := translator.Trans("TEST_PARAMETERS", "Lev", 2)
	want := "Hello Lev. You are user number 2 with username Lev"
	if got != want {
		t.Errorf("renderer %s != %s", got, want)
	}
}
