package renderer

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v3"
)

type RendererConfig struct {
	TemplatesFolder  string
	TranslationsFile string
}

type Renderer struct {
	translationMap map[string]interface{}
	config         *RendererConfig
}

// NewRenderer will load a translation file and return a new template renderer.
func NewRenderer(config *RendererConfig) (*Renderer, error) {
	translation, err := ioutil.ReadFile(config.TranslationsFile)
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
		config:         config,
	}

	return r, nil
}

// Render a template with the specified PageView data.
func (r *Renderer) RenderTemplate(w io.Writer, t string, v interface{}) {
	var templates = template.Must(
		template.New("").Funcs(template.FuncMap{
			"trans":      r.trans,
			"total_time": r.totalTime,
		}).ParseGlob(r.config.TemplatesFolder + "/*.html"))

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

// Returns the difference between the start time and the end time
func (r *Renderer) totalTime(start, end time.Time) time.Duration {
	return end.Sub(start)
}
