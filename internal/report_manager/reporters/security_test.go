package reporters_test

import (
	"net/http"
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporters"

	"golang.org/x/net/html"
)

// Test the MissingHSTSHeader reporter with a PageReport with HSTS header.
// The reporter should not report the issue.
func TestMissingHSTSHeaderIssues(t *testing.T) {
	reporter := reporters.NewMissingHSTSHeaderReporter()
	if reporter.ErrorType != reporter_errors.ErrorMissingHSTSHeader {
		t.Errorf("TestMissingHSTSHeaderIssues: error type is not correct")
	}

	// Run the reporter callback with the PageReport.
	reportsIssue := reporter.Callback(&models.PageReport{}, &html.Node{}, &http.Header{})

	// The reporter should not found any issue.
	if reportsIssue == false {
		t.Errorf("TestMissingHSTSHeaderIssues: reportsIssue should be true")
	}
}

// Test the MissingHSTSHeader reporter with a PageReport with HSTS header.
// The reporter should not report the issue.
func TestMissingHSTSHeaderNoIssues(t *testing.T) {
	reporter := reporters.NewMissingHSTSHeaderReporter()
	if reporter.ErrorType != reporter_errors.ErrorMissingHSTSHeader {
		t.Errorf("TestMissingHSTSHeaderNoIssues: error type is not correct")
	}

	header := &http.Header{}
	header.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

	// Run the reporter callback with the PageReport.
	reportsIssue := reporter.Callback(&models.PageReport{}, &html.Node{}, header)

	// The reporter should not found any issue.
	if reportsIssue == true {
		t.Errorf("TestMissingHSTSHeaderNoIssues: reportsIssue should be false")
	}
}
