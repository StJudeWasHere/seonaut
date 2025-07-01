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
		Trans(s string, args ...interface{}) string
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

	var err error
	r.templates, err = findAndParseTemplates(config.TemplatesFolder, template.FuncMap{
		"trans":      r.translator.Trans,
		"total_time": r.totalTime,
		"add":        r.add,
		"to_kb":      r.toKByte,
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
