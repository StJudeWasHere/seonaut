package page

import (
	"net/http"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
)

// Returns a report_manager.PageIssueReporter with a callback function that
// checks if a page has a high depth. The callback returns true if the page is text/html,
// has a 20x status code and has high depth.
func NewDepthReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.StatusCode < 200 || pageReport.StatusCode >= 300 {
			return false
		}

		return pageReport.Depth > 4
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorDepth,
		Callback:  c,
	}
}
