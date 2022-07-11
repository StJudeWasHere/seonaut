package export

import (
	"encoding/csv"
	"io"

	"github.com/stjudewashere/seonaut/internal/crawler"
)

type Link struct {
	Origin      string
	Destination string
	Text        string
}

type Store interface {
	ExportLinks(*crawler.Crawl) <-chan *Link
	ExportExternalLinks(*crawler.Crawl) <-chan *Link
}
type Exporter struct {
	store Store
}

func NewExporter(s Store) *Exporter {
	return &Exporter{
		store: s,
	}
}

// Export internal links as a CSV file
func (e *Exporter) ExportLinks(f io.Writer, crawl *crawler.Crawl) {
	w := csv.NewWriter(f)

	w.Write([]string{
		"Origin",
		"Destination",
		"Text",
	})

	lStream := e.store.ExportLinks(crawl)

	for v := range lStream {
		w.Write([]string{
			v.Origin,
			v.Destination,
			v.Text,
		})
	}

	w.Flush()
}

// Export internal links as a CSV file
func (e *Exporter) ExportExternalLinks(f io.Writer, crawl *crawler.Crawl) {
	w := csv.NewWriter(f)

	w.Write([]string{
		"Origin",
		"Destination",
		"Text",
	})

	lStream := e.store.ExportExternalLinks(crawl)

	for v := range lStream {
		w.Write([]string{
			v.Origin,
			v.Destination,
			v.Text,
		})
	}

	w.Flush()
}
