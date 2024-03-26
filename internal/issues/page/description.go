package page

import (
	"net/http"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// Returns a report_manager.PageIssueReporter with a callback function that checks if a page has
// an empty or missing description. It returns true if the status code is between
// 200 and 299, the media type is text/html and the description is not set.
func NewEmptyDescriptionReporter() *models.PageIssueReporter {
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

		return pageReport.Description == ""
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorEmptyDescription,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that checks if a page has a short description.
// The callback function returns true if the page is text/html, has a status code between 200 and 299,
// and has a description of less than an specified amount of letters.
func NewShortDescriptionReporter() *models.PageIssueReporter {
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

		return len(pageReport.Description) > 0 && len(pageReport.Description) < 80
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorShortDescription,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that checks if a page has a short description.
// The callback function returns true if the page is text/html, has a status code between 200 and 299,
// and has a description of more than an specified amount of letters.
func NewLongDescriptionReporter() *models.PageIssueReporter {
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

		return len(pageReport.Description) > 160
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorLongDescription,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that checks if the page has more
// than one description meta tag in the header section.
// The callback returns true if the page is text/html and has more than one description in the header section.
func NewMultipleDescriptionTagsReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		tags, err := htmlquery.QueryAll(htmlNode, "//head//meta[@name=\"description\"]")
		if err != nil {
			return false
		}

		return len(tags) > 1
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorMultipleDescriptionTags,
		Callback:  c,
	}
}
