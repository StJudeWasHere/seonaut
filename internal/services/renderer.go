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
)

type (
	RendererTranslator interface {
		Trans(lang, s string, args ...interface{}) string
		TransDate(lang string, d time.Time, f string) string
	}

	RendererConfig struct {
		TemplatesFolder string
	}

	Renderer struct {
		config     *RendererConfig
		templates  *template.Template
		translator RendererTranslator
	}
)

// NewRenderer returns a new template renderer with the specified configuration.
func NewRenderer(config *RendererConfig, translator RendererTranslator) (*Renderer, error) {
	r := &Renderer{
		translator: translator,
		config:     config,
	}

	commonFuncMap := template.FuncMap{
		"total_time": r.totalTime,
		"add":        r.add,
		"to_kb":      r.toKByte,
		"is_rtl":     r.isRTL,
		"trans":      func(s string, args ...interface{}) string { return s },   // This gets replaced with the user lang in the RenderTemplate
		"trans_date": func(d time.Time, f string) string { return d.Format(f) }, // This gets replaced with the user lang in the RenderTemplate
	}

	var err error
	r.templates, err = findAndParseTemplates(config.TemplatesFolder, commonFuncMap)
	if err != nil {
		return nil, fmt.Errorf("renderer initialisation failed: %w", err)
	}

	return r, nil
}

// Render a template with the specified PageView data.
func (r *Renderer) RenderTemplate(w io.Writer, t string, v interface{}, lang string) {
	// Replace the translation functions so they use the translator with the user language.
	funcMap := template.FuncMap{
		"trans": func(s string, args ...interface{}) string {
			return r.translator.Trans(lang, s, args...)
		},
		"trans_date": func(d time.Time, f string) string {
			return r.translator.TransDate(lang, d, f)
		},
	}

	tmpl := template.Must(r.templates.Clone())
	tmpl.Funcs(funcMap)

	err := tmpl.ExecuteTemplate(w, t+".html", v)
	if err != nil {
		log.Printf("RenderTemplate: %v\n", err)
	}
}

// Returns the difference between the start time and the end time.
func (r *Renderer) totalTime(start, end time.Time) time.Duration {
	return end.Sub(start)
}

// Helper function to add integers in the templates.
func (r *Renderer) add(i ...int) int {
	total := 0
	for _, v := range i {
		total += v
	}

	return total
}

// isRTL takes a language code as a parameter and returns true if it is a RTL language.
// Otherwise it returns false.
func (r *Renderer) isRTL(lang string) bool {
	rtlLangs := []string{
		"ar",
		"he",
		"fa",
		"ur",
	}

	for _, l := range rtlLangs {
		if lang == l {
			return true
		}
	}

	return false
}

// toKByte is a helper function that returns an int64 formated as KB.
func (r *Renderer) toKByte(b int64) string {
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
