package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
)

// Returns a new PageIssueReporter with a callback function that
// checks if the status code is in the 30x range.
func NewStatus30xReporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		if pageReport.Crawled == false {
			return false
		}

		return pageReport.StatusCode >= 300 && pageReport.StatusCode < 400
	}

	return &PageIssueReporter{
		ErrorType: Error30x,
		Callback:  c,
	}
}

// Returns a new PageIssueReporter with a callback function that
// checks if the status code is in the 40x range.
func NewStatus40xReporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		if pageReport.Crawled == false {
			return false
		}

		return pageReport.StatusCode >= 400 && pageReport.StatusCode < 500
	}

	return &PageIssueReporter{
		ErrorType: Error40x,
		Callback:  c,
	}
}

// Returns a new PageIssueReporter with a callback function that
// checks if the status code is greater or equal than 500.
func NewStatus50xReporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		if pageReport.Crawled == false {
			return false
		}

		return pageReport.StatusCode >= 500
	}

	return &PageIssueReporter{
		ErrorType: Error50x,
		Callback:  c,
	}
}
