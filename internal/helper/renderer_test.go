package helper_test

import (
	"bytes"
	"testing"

	"github.com/stjudewashere/seonaut/internal/helper"
)

func TestRenderer(t *testing.T) {
	renderer, err := helper.NewRenderer(&helper.RendererConfig{
		TemplatesFolder:  "./testdata",
		TranslationsFile: "./testdata/translations.test.yaml",
	})
	if err != nil {
		t.Errorf("%v", err)
	}

	eb := new(bytes.Buffer)
	e := "Page Title: Test Title"

	renderer.RenderTemplate(eb, "test", &helper.PageView{PageTitle: "Test Title"})
	if eb.String() != e {
		t.Errorf("renderer %s != %s", eb.String(), e)
	}

	tb := new(bytes.Buffer)
	te := "test translation"

	renderer.RenderTemplate(tb, "translations", &helper.PageView{})
	if tb.String() != te {
		t.Errorf("renderer %s != %s", tb.String(), te)
	}
}
