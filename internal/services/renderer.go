package services

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type (
	RendererConfig struct {
		TemplatesFolder  string
		TranslationsFile string
	}

	Renderer struct {
		translationMap map[string]interface{}
		config         *RendererConfig
		templates      *template.Template
	}
)

// NewRenderer will load a translation file and return a new template renderer.
func NewRenderer(config *RendererConfig) (*Renderer, error) {
	translation, err := os.ReadFile(config.TranslationsFile)
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

	r.templates, err = findAndParseTemplates(config.TemplatesFolder, template.FuncMap{
		"trans":      r.trans,
		"total_time": r.totalTime,
		"add":        r.add,
		"to_kb":      r.ToKByte,
	})
	if err != nil {
		return nil, fmt.Errorf("renderer initialisation failed: %w", err)
	}

	return r, nil
}

// Render a template with the specified PageView data.
func (r *Renderer) RenderTemplate(w io.Writer, t string, v interface{}) {
	err := r.templates.ExecuteTemplate(w, t+".html", v)
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

// Add integers
func (r *Renderer) add(i ...int) int {
	total := 0
	for _, v := range i {
		total += v
	}

	return total
}

// Returns an int formated as KB.
func (r *Renderer) ToKByte(b int64) string {
	v := b / (1 << 10)
	i := b % (1 << 10)

	kb := float64(v) + float64(i)/float64(1<<10)
	formatted := fmt.Sprintf("%.2f", kb)

	return formatted
}

// findAndParseTemplates locates and parses all HTML template files in the specified directory.
// It returns the Template object and an error if any issues occur during parsing.
func findAndParseTemplates(rootDir string, funcMap template.FuncMap) (*template.Template, error) {
	cleanRoot := filepath.Clean(rootDir)
	pfx := len(cleanRoot) + 1
	root := template.New("")

	err := filepath.Walk(cleanRoot, func(path string, info os.FileInfo, e1 error) error {
		if !info.IsDir() && strings.HasSuffix(path, ".html") {
			if e1 != nil {
				return fmt.Errorf("file walk error: %w", e1)
			}

			b, e2 := os.ReadFile(path)
			if e2 != nil {
				return fmt.Errorf("read file %s error: %w", path, e2)
			}

			name := path[pfx:]
			t := root.New(name).Funcs(funcMap)
			_, e2 = t.Parse(string(b))
			if e2 != nil {
				return fmt.Errorf("parse template %s error: %w", name, e2)
			}
		}

		return nil
	})

	return root, err
}
