package services_test

import (
	"testing"

	"github.com/stjudewashere/seonaut/internal/services"
)

// Test the translator's Trans method
func TestTrans(t *testing.T) {
	lang := "en"
	translator, err := services.NewTranslator("./testdata", lang)
	if err != nil {
		t.Fatalf("%v", err)
	}

	got := translator.Trans(lang, "TEST")
	want := "test translation"
	if got != want {
		t.Errorf("renderer %s != %s", got, want)
	}
}

// Test the translator's Trans method with parameters
func TestTransParameters(t *testing.T) {
	lang := "en"
	translator, err := services.NewTranslator("./testdata", lang)
	if err != nil {
		t.Fatalf("%v", err)
	}

	got := translator.Trans(lang, "TEST_PARAMETERS", "Lev", 2)
	want := "Hello Lev. You are user number 2 with username Lev"
	if got != want {
		t.Errorf("renderer %s != %s", got, want)
	}
}
