package page

import (
	"net/http"
	"strings"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
)

// Returns a report_manager.PageIssueReporter with a callback function that checks
// if URL has undescore characters.
func NewUnderscoreURLReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		return strings.Contains(pageReport.URL, "_")
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorUnderscoreURL,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that checks
// if URL has a space characters.
func NewSpaceURLReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		return strings.Contains(pageReport.URL, " ")
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorSpaceURL,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that checks
// if URL has multiple slash characters.
func NewMultipleSlashesReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		return strings.Contains(pageReport.ParsedURL.Path, "//")
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorMultipleSlashes,
		Callback:  c,
	}
}
