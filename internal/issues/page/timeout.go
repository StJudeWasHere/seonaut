package page

import (
	"net/http"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
)

// Returns a report_manager.PageIssueReporter with a callback function that
// checks if a web page timedout. The callback returns true if the page timed out.
func NewTimeoutReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		return pageReport.Timeout
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorTimeout,
		Callback:  c,
	}
}
