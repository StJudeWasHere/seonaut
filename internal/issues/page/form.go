package page

import (
	"net/http"
	"net/url"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// contains a form on an insecure URL.
func NewFormOnHTTPReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if pageReport.ParsedURL.Scheme == "https" {
			return false
		}

		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.StatusCode < 200 && pageReport.StatusCode >= 300 {
			return false
		}

		forms, err := htmlquery.QueryAll(htmlNode, "//form")
		if err != nil || forms == nil {
			return false
		}

		return true
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorFormOnHTTP,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// contains a form with an insecure action URL.
func NewInsecureFormReporter() *models.PageIssueReporter {
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

		forms, err := htmlquery.QueryAll(htmlNode, "//form")
		if err != nil || forms == nil {
			return false
		}

		for _, f := range forms {
			action := htmlquery.SelectAttr(f, "action")
			u, err := url.Parse(action)
			if err != nil {
				continue
			}

			if u.Scheme == "http" {
				return true
			}
		}

		return false
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorInsecureForm,
		Callback:  c,
	}
}
