package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Returns a PageIssueReporter with a callback function to check
// if a page has images with no alt attribute. The callback returns true in case
// the page is text/html and contains images with empty or missing alt attribute.
func NewAltTextReporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		if pageReport.Crawled == false {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		for _, i := range pageReport.Images {
			if i.Alt == "" {
				return true
			}
		}

		return false
	}

	return &PageIssueReporter{
		ErrorType: reporter_errors.ErrorImagesWithNoAlt,
		Callback:  c,
	}
}
