package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"unicode/utf8"

	"github.com/stjudewashere/seonaut/internal/models"
)

type CSVWriter struct {
	writer *csv.Writer
}

func NewCSVWriter(f io.Writer) *CSVWriter {
	cw := CSVWriter{}
	cw.writer = csv.NewWriter(f)

	cw.writer.Write([]string{
		"Status Code",
		"URL",
		"Redirect URL",
		"Content Type",
		"Canonical",
		"Lang",
		"Title",
		"Title Length",
		"Description",
		"Description Length",
		"Robots",
		"Header 1",
		"Header 2",
		"Size",
		"Nº of words",
	})

	return &cw
}

func (cw *CSVWriter) Write(r *models.PageReport) {
	cw.writer.Write([]string{
		fmt.Sprintf("%d", r.StatusCode),
		r.URL,
		r.RedirectURL,
		r.ContentType,
		r.Canonical,
		r.Lang,
		r.Title,
		fmt.Sprint(utf8.RuneCount([]byte(r.Title))),
		r.Description,
		fmt.Sprint(utf8.RuneCount([]byte(r.Description))),
		r.Robots,
		r.H1,
		r.H2,
		fmt.Sprintf("%.1f KB", byteToKByte(r.Size)),
		strconv.Itoa(r.Words),
	})

	cw.writer.Flush()
}

func byteToKByte(b int64) float64 {
	v := b / (1 << 10)
	r := b % (1 << 10)

	return float64(v) + float64(r)/float64(1<<10)
}
