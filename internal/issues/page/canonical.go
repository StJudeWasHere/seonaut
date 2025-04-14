package page

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// head contains more than one canonical tag.
func NewCanonicalMultipleTagsReporter() *models.PageIssueReporter {
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

		canonical, err := htmlquery.QueryAll(htmlNode, "//head/link[@rel=\"canonical\"]/@href")
		if err != nil || canonical == nil {
			return false
		}

		return len(canonical) > 1
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorMultipleCanonicalTags,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// canonical tag is using a relative URL.
func NewCanonicalRelativeURLReporter() *models.PageIssueReporter {
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

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorRelativeCanonicalURL,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// head canonical tag and the canonical header don't match.
func NewCanonicalMismatchReporter() *models.PageIssueReporter {
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

		link, err := htmlquery.Query(htmlNode, "//head/link[@rel=\"canonical\"]/@href")
		if err != nil || link == nil {
			return false
		}

		tagCanonical := htmlquery.SelectAttr(link, "href")

		headerCanonical := ""
		linkHeaderElements := strings.Split(header.Get("Link"), ",")
		for _, lh := range linkHeaderElements {
			attr := strings.Split(lh, ";")
			if len(attr) == 2 && strings.Contains(attr[1], `rel="canonical"`) {
				canonicalString := strings.TrimSpace(attr[0])
				headerCanonical = canonicalString[1 : len(canonicalString)-1]
			}
		}

		if headerCanonical == "" {
			return false
		}

		return tagCanonical != headerCanonical
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorCanonicalMismatch,
		Callback:  c,
	}
}
