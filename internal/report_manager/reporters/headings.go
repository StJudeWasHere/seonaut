package reporters

import (
	"golang.org/x/net/html"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// doesn't have any H1 tag.
func NewNoH1Reporter() *report_manager.PageIssueReporter {
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

		return pageReport.H1 == ""
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorNoH1,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the heading tags
// in the page's html doesn't have the correct order.
func NewValidHeadingsOrderReporter() *report_manager.PageIssueReporter {
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

		if pageReport.ValidHeadings {
			return false
		}

		return true
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorNotValidHeadings,
		Callback:  c,
	}
}
