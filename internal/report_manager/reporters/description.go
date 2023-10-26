package reporters

import (
	"net/http"

	"golang.org/x/net/html"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Returns a report_manager.PageIssueReporter with a callback function that checks if a page has
// an empty or missing description. It returns true if the status code is between
// 200 and 299, the media type is text/html and the description is not set.
func NewEmptyDescriptionReporter() *report_manager.PageIssueReporter {
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

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorEmptyDescription,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that checks if a page has a short description.
// The callback function returns true if the page is text/html, has a status code between 200 and 299,
// and has a description of less than an specified amount of letters.
func NewShortDescriptionReporter() *report_manager.PageIssueReporter {
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

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorShortDescription,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that checks if a page has a short description.
// The callback function returns true if the page is text/html, has a status code between 200 and 299,
// and has a description of more than an specified amount of letters.
func NewLongDescriptionReporter() *report_manager.PageIssueReporter {
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

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorLongDescription,
		Callback:  c,
	}
}
