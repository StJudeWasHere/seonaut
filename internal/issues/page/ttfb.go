package page

import (
	"net/http"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
)

// Returns a report_manager.PageIssueReporter with a callback function that
// checks if the TTFB. The callback returns true if the page's time to first byte is slow.
func NewSlowTTFBReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		return pageReport.TTFB > 800
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorSlowTTFB,
		Callback:  c,
	}
}
