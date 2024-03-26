package page

import (
	"net/http"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
)

// Returns a report_manager.PageIssueReporter with a callback function that checks if page uses the http
// scheme instead of https. The callback function returns true has a 20x status code and uses http scheme.
func NewHTTPSchemeReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.StatusCode < 200 || pageReport.StatusCode >= 300 {
			return false
		}

		return pageReport.ParsedURL.Scheme == "http"
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorHTTPScheme,
		Callback:  c,
	}
}
