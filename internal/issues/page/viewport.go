package page

import (
	"net/http"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// head does not contain a viewport meta tag or if the meta viewport tag content is empty.
func NewViewportTagReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.StatusCode < 200 && pageReport.StatusCode >= 300 {
			return false
		}

		viewport, err := htmlquery.Query(htmlNode, "//head/meta[@name=\"viewport\"]")
		if err != nil || viewport == nil {
			return true
		}

		viewportContent := htmlquery.SelectAttr(viewport, "content")
		if viewportContent == "" {
			return true
		}

		return false
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorMissingViewportTag,
		Callback:  c,
	}
}
