package reporters

import (
	"net/http"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/text/language"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the status code media type is text/html and the page's html language is not valid.
func NewInvalidLangReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.Lang == "" {
			return false
		}

		langs := strings.Split(pageReport.Lang, ",")
		for _, l := range langs {
			_, err := language.Parse(l)
			if err != nil {
				return true
			}
		}

		return false
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorInvalidLanguage,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the status code media type is text/html and the page's html language is missing or empty.
func NewMissingLangReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
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
