package reporters

import (
	"golang.org/x/net/html"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the status code media type is text/html and the page's html language is not valid.
func NewInvalidLangReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.ValidLang {
			return false
		}

		return true
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorInvalidLanguage,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the status code media type is text/html and the page's html language is missing or empty.
func NewMissingLangReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		return pageReport.Lang == ""
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorNoLang,
		Callback:  c,
	}
}
