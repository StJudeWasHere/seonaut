package page

import (
	"net/http"
	"strings"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
	"golang.org/x/text/language"
)

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the status code media type is text/html and the page's html language is not valid.
func NewInvalidLangReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.StatusCode >= 300 && pageReport.StatusCode < 400 {
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

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorInvalidLanguage,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the status code media type is text/html and the page's html language is missing or empty.
func NewMissingLangReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.StatusCode >= 300 && pageReport.StatusCode < 400 {
			return false
		}

		return pageReport.Lang == ""
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorNoLang,
		Callback:  c,
	}
}
