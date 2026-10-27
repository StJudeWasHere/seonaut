package services

import (
	"bytes"
	"encoding/csv"
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
)

// TestExportPageReportsIncludesLinkStats checks that the pagereports CSV
// export includes the incoming links count, an example origin URL and the
// outgoing links count for every exported page report.
func TestExportPageReportsIncludesLinkStats(t *testing.T) {
	e := NewExporter(nil, nil)

	prStream := make(chan *models.PageReport, 1)
	prStream <- &models.PageReport{
		URL:           "https://example.com/redirect",
		RedirectURL:   "https://example.com/final",
		StatusCode:    301,
		IncomingLinks: 5820,
		ExampleOrigin: "https://example.com/",
		OutgoingLinks: 5,
	}
	close(prStream)

	var buf bytes.Buffer
	e.ExportPageReports(&buf, prStream)

	rows, err := csv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatalf("failed to parse csv output: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("expected 2 csv rows (header + data), got %d", len(rows))
	}

	// Locate the new columns in the header row.
	col := map[string]int{
		"Incoming Links": -1,
		"Example Origin": -1,
		"Outgoing Links": -1,
	}
	for i, h := range rows[0] {
		if _, ok := col[h]; ok {
			col[h] = i
		}
	}
	for h, i := range col {
		if i == -1 {
			t.Errorf("csv header is missing column %q", h)
		}
	}

	data := rows[1]
	if got, want := data[col["Incoming Links"]], "5820"; got != want {
		t.Errorf("Incoming Links = %q, want %q", got, want)
	}
	if got, want := data[col["Example Origin"]], "https://example.com/"; got != want {
		t.Errorf("Example Origin = %q, want %q", got, want)
	}
	if got, want := data[col["Outgoing Links"]], "5"; got != want {
		t.Errorf("Outgoing Links = %q, want %q", got, want)
	}
}

// TestExportPageReportsZeroLinkStats checks the link stats columns are
// written as zero/empty when the page report has no link data.
func TestExportPageReportsZeroLinkStats(t *testing.T) {
	e := NewExporter(nil, nil)

	prStream := make(chan *models.PageReport, 1)
	prStream <- &models.PageReport{
		URL:        "https://example.com/no-links",
		StatusCode: 200,
	}
	close(prStream)

	var buf bytes.Buffer
	e.ExportPageReports(&buf, prStream)

	rows, err := csv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatalf("failed to parse csv output: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("expected 2 csv rows (header + data), got %d", len(rows))
	}

	incomingIdx, exampleIdx, outgoingIdx := -1, -1, -1
	for i, h := range rows[0] {
		switch h {
		case "Incoming Links":
			incomingIdx = i
		case "Example Origin":
			exampleIdx = i
		case "Outgoing Links":
			outgoingIdx = i
		}
	}
	if incomingIdx == -1 || exampleIdx == -1 || outgoingIdx == -1 {
		t.Fatalf("csv header is missing link stats columns")
	}

	if got := rows[1][incomingIdx]; got != "0" {
		t.Errorf("Incoming Links = %q, want %q", got, "0")
	}
	if got := rows[1][exampleIdx]; got != "" {
		t.Errorf("Example Origin = %q, want empty", got)
	}
	if got := rows[1][outgoingIdx]; got != "0" {
		t.Errorf("Outgoing Links = %q, want %q", got, "0")
	}
}
