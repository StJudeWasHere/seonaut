package reporters

import (
	"net/http"

	"golang.org/x/net/html"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Returns a report_manager.PageIssueReporter with a callback function that checks if page uses the http
// scheme instead of https. The callback function returns true has a 20x status code and uses http scheme.
func NewHTTPSchemeReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.StatusCode < 200 || pageReport.StatusCode >= 300 {
			return false
		}

		return pageReport.ParsedURL.Scheme == "http"
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorHTTPScheme,
		Callback:  c,
	}
}
