package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Returns a PageIssueReporter with a callback function that returns true if
// the status code media type is text/html and the page's html language is not valid.
func NewInvalidLangReporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		if pageReport.Crawled == false {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.ValidLang == true {
			return false
		}

		return true
	}

	return &PageIssueReporter{
		ErrorType: reporter_errors.ErrorInvalidLanguage,
		Callback:  c,
	}
}

// Returns a PageIssueReporter with a callback function that returns true if
// the status code media type is text/html and the page's html language is missing or empty.
func NewMissingLangReporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		if pageReport.Crawled == false {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		return pageReport.Lang == ""
	}

	return &PageIssueReporter{
		ErrorType: reporter_errors.ErrorNoLang,
		Callback:  c,
	}
}
