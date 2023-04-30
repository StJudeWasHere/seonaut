package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
)

// Returns a PageIssueReporter with a callback function that checks if page has an empty little.
// The callback function returns true if the page is text/html, has a 20x status code
// and has an empty or missing title.
func NewEmptyTitleReporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		if pageReport.Crawled == false {
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

	return &PageIssueReporter{
		ErrorType: ErrorEmptyTitle,
		Callback:  c,
	}
}

// Returns a PageIssueReporter with a callback function that checks if the page has a short title.
// The callback returns true if the page is text/html and has a page title shorter than an specified
// amount of letters.
func NewShortTitleReporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		if pageReport.Crawled == false {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		return len(pageReport.Title) > 0 && len(pageReport.Title) < 20
	}

	return &PageIssueReporter{
		ErrorType: ErrorShortTitle,
		Callback:  c,
	}
}

// Returns a PageIssueReporter with a callback function that checks if the page has a long title.
// The callback function returns true if the page is text/html and has a page title longer than an
// specified amount of letters.
func NewLongTitleReporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		if pageReport.Crawled == false {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		return len(pageReport.Title) > 60
	}

	return &PageIssueReporter{
		ErrorType: ErrorLongTitle,
		Callback:  c,
	}
}
