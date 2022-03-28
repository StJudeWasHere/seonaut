package helper

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/stjudewashere/seonaut/internal/user"

	"gopkg.in/yaml.v3"
)

type PageView struct {
	PageTitle string
	User      user.User
	Data      interface{}
	Refresh   bool
}

type Renderer struct {
	translationMap map[string]interface{}
}

// NewRenderer will load a translation file and return a new template renderer.
func NewRenderer() (*Renderer, error) {
	translation, err := ioutil.ReadFile("translations/translation.en.yaml")
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	err = yaml.Unmarshal(translation, &m)
	if err != nil {
		return nil, err
	}

	r := &Renderer{
		translationMap: m,
	}

	return r, nil
}

// Render a template with the specified PageView data.
func (r *Renderer) RenderTemplate(w http.ResponseWriter, t string, v *PageView) {
	var templates = template.Must(
		template.New("").Funcs(template.FuncMap{
			"trans": r.trans,
		}).ParseGlob("web/templates/*.html"))

	err := templates.ExecuteTemplate(w, t+".html", v)
	if err != nil {
		log.Printf("RenderTemplate: %v\n", err)
	}
}

// Returns a translated string from the translations map.
// The original string is returned if a translation is not found in the map.
func (r *Renderer) trans(s string) string {
	t, ok := r.translationMap[s]
	if !ok {
		log.Printf("trans: %s translation not found\n", s)
		return s
	}

	return fmt.Sprintf("%v", t)
}
