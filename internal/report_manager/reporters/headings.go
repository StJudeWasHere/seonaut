package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Returns a PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// doesn't have any H1 tag.
func NewNoH1Reporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		if pageReport.Crawled == false {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.StatusCode < 200 && pageReport.StatusCode >= 300 {
			return false
		}

		return pageReport.H1 == ""
	}

	return &PageIssueReporter{
		ErrorType: reporter_errors.ErrorNoH1,
		Callback:  c,
	}
}

// Returns a PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the heading tags
// in the page's html doesn't have the correct order.
func NewValidHeadingsOrderReporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		if pageReport.Crawled == false {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.StatusCode < 200 && pageReport.StatusCode >= 300 {
			return false
		}

		if pageReport.ValidHeadings == true {
			return false
		}

		return true
	}

	return &PageIssueReporter{
		ErrorType: reporter_errors.ErrorNotValidHeadings,
		Callback:  c,
	}
}
