package reporters

import (
	"net/http"

	"golang.org/x/net/html"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Returns a report_manager.PageIssueReporter with a callback function that checks if page has an empty little.
// The callback function returns true if the page is text/html, has a 20x status code
// and has an empty or missing title.
func NewEmptyTitleReporter() *report_manager.PageIssueReporter {
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

		return pageReport.Title == ""
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorEmptyTitle,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that checks if the page has a short title.
// The callback returns true if the page is text/html and has a page title shorter than an specified
// amount of letters.
func NewShortTitleReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		return len(pageReport.Title) > 0 && len(pageReport.Title) < 20
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorShortTitle,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that checks if the page has a long title.
// The callback function returns true if the page is text/html and has a page title longer than an
// specified amount of letters.
func NewLongTitleReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		return len(pageReport.Title) > 60
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorLongTitle,
		Callback:  c,
	}
}
