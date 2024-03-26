package page

import (
	"net/http"

	"golang.org/x/net/html"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"
)

// Returns a report_manager.PageIssueReporter with a callback function that
// checks if a page has little content. The callback returns true if the page is text/html,
// has a 20x status code and less than a specified amount of words.
func NewLittleContentReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.StatusCode < 200 || pageReport.StatusCode >= 300 {
			return false
		}

		return pageReport.Words < 200
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorLittleContent,
		Callback:  c,
	}
}
