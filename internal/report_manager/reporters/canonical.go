package reporters

import (
	"net/url"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// head contains more than one canonical tag.
func NewCanonicalMultipleTagsReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.StatusCode < 200 && pageReport.StatusCode >= 300 {
			return false
		}

		canonical, err := htmlquery.QueryAll(htmlNode, "//head/link[@rel=\"canonical\"]/@href")
		if err != nil || canonical == nil {
			return false
		}

		return len(canonical) > 1
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorMultipleCanonicalTags,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// canonical tag is using a relative URL.
func NewCanonicalRelativeURLReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.StatusCode < 200 && pageReport.StatusCode >= 300 {
			return false
		}

		canonical, err := htmlquery.Query(htmlNode, "//head/link[@rel=\"canonical\"]/@href")
		if err != nil || canonical == nil {
			return false
		}

		parsedURL, err := url.Parse(htmlquery.SelectAttr(canonical, "href"))
		if err != nil {
			return false
		}

		return !parsedURL.IsAbs()
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorRelativeCanonicalURL,
		Callback:  c,
	}
}
