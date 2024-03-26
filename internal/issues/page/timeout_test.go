package page_test

import (
	"net/http"
	"testing"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/issues/page"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
)

// Test the Depth reporter with a pageReport that does not
// have a a depth issue. The reporter should not report the issue.
func TestTimeoutNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := page.NewTimeoutReporter()
	if reporter.ErrorType != errors.ErrorTimeout {
		t.Errorf("error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("reportsIssue should be false")
	}
}

// Test the Depth reporter with a pageReport that does
// have a depth issue. The reporter should report the issue.
func TestTimeoutIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Timeout: true,
	}

	reporter := page.NewTimeoutReporter()
	if reporter.ErrorType != errors.ErrorTimeout {
		t.Errorf("error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}
}
