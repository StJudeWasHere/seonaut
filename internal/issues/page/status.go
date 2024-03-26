package page

import (
	"net/http"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
)

// Returns a new report_manager.PageIssueReporter with a callback function that
// checks if the status code is in the 30x range.
func NewStatus30xReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		return pageReport.StatusCode >= 300 && pageReport.StatusCode < 400
	}

	return &models.PageIssueReporter{
		ErrorType: errors.Error30x,
		Callback:  c,
	}
}

// Returns a new report_manager.PageIssueReporter with a callback function that
// checks if the status code is in the 40x range.
func NewStatus40xReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		return pageReport.StatusCode >= 400 && pageReport.StatusCode < 500
	}

	return &models.PageIssueReporter{
		ErrorType: errors.Error40x,
		Callback:  c,
	}
}

// Returns a new report_manager.PageIssueReporter with a callback function that
// checks if the status code is greater or equal than 500.
func NewStatus50xReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		return pageReport.StatusCode >= 500
	}

	return &models.PageIssueReporter{
		ErrorType: errors.Error50x,
		Callback:  c,
	}
}
