package container_test

import (
	"bytes"
	"testing"

	"github.com/stjudewashere/seonaut/internal/container"
)

func TestRenderer(t *testing.T) {
	r, err := container.NewRenderer(&container.RendererConfig{
		TemplatesFolder:  "./testdata",
		TranslationsFile: "./testdata/translations.test.yaml",
	})
	if err != nil {
		t.Fatalf("%v", err)
	}

	eb := new(bytes.Buffer)
	e := "Page Title: Test Title"

	r.RenderTemplate(eb, "test", &struct{ PageTitle string }{PageTitle: "Test Title"})
	if eb.String() != e {
		t.Errorf("renderer %s != %s", eb.String(), e)
	}

	tb := new(bytes.Buffer)
	te := "test translation"

	r.RenderTemplate(tb, "translations", &struct{}{})
	if tb.String() != te {
		t.Errorf("renderer %s != %s", tb.String(), te)
	}
}
