package reporters

import (
	"net/http"

	"golang.org/x/net/html"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Returns a report_manager.PageIssueReporter with a callback function that
// checks if a web page timedout. The callback returns true if the page timed out.
func NewTimeoutReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		return pageReport.Timeout
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorTimeout,
		Callback:  c,
	}
}
