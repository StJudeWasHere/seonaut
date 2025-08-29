package services_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stjudewashere/seonaut/internal/services"
)

type TestTranslator struct{}

func (t *TestTranslator) Trans(lang, s string, args ...interface{}) string {
	return s
}

func (t *TestTranslator) TransDate(lang string, d time.Time, f string) string {
	return d.Format(f)
}

func TestRenderer(t *testing.T) {
	r, err := services.NewRenderer(&services.RendererConfig{
		TemplatesFolder: "./testdata",
	}, &TestTranslator{})
	if err != nil {
		t.Fatalf("%v", err)
	}

	eb := new(bytes.Buffer)
	e := "Page Title: Test Title"
	lang := "en"
	r.RenderTemplate(eb, "test", &struct{ PageTitle string }{PageTitle: "Test Title"}, lang)
	if eb.String() != e {
		t.Errorf("renderer %s != %s", eb.String(), e)
	}
}
