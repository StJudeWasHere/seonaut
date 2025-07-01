package services_test

import (
	"bytes"
	"testing"

	"github.com/stjudewashere/seonaut/internal/services"
)

type TestTranslator struct{}

func (t *TestTranslator) Trans(s string, args ...interface{}) string {
	return s
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

	r.RenderTemplate(eb, "test", &struct{ PageTitle string }{PageTitle: "Test Title"})
	if eb.String() != e {
		t.Errorf("renderer %s != %s", eb.String(), e)
	}
}
