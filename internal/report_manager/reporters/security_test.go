package reporters_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporters"

	"golang.org/x/net/html"
)

// Test the MissingHSTSHeader reporter with a PageReport with HSTS header.
// The reporter should report the issue.
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

// Test the MissingCSP reporter with a PageReport with CSP.
// The reporter should not report the issue.
func TestMissingCSPNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := reporters.NewMissingCSPReporter()
	if reporter.ErrorType != reporter_errors.ErrorMissingCSP {
		t.Errorf("TestMissingCSPIssues: error type is not correct")
	}

	source := `
		<html>
			<head>
				<meta http-equiv="Content-Security-Policy" content="default-src 'self'">
			</head>
		</html>
	`

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("error parsing html source")
	}

	// Run the reporter callback with the PageReport.
	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	// The reporter should not found any issue.
	if reportsIssue == true {
		t.Errorf("reportsIssue should be false")
	}

	source = `
	<html>
		<head>
		</head>
	</html>
`

	doc, err = html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("error parsing html source")
	}

	// Check if CSP is set in the header.
	header := &http.Header{}
	header.Set("Content-Security-Policy", "default-src 'self'")

	reportsIssue = reporter.Callback(pageReport, doc, header)

	// The reporter should not found any issue.
	if reportsIssue == true {
		t.Errorf("reportsIssue should be false")
	}
}

// Test the MissingCSP reporter with a PageReport without CSP.
// The reporter should report the issue.
func TestMissingCSPIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := reporters.NewMissingCSPReporter()
	if reporter.ErrorType != reporter_errors.ErrorMissingCSP {
		t.Errorf("TestMissingCSPIssues: error type is not correct")
	}

	source := `
		<html>
			<head>
			</head>
		</html>
	`

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("error parsing html source")
	}

	// Run the reporter callback with the PageReport.
	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	// The reporter should not found any issue.
	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}

	header := &http.Header{}

	reportsIssue = reporter.Callback(pageReport, doc, header)

	// The reporter should not found any issue.
	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}
}

// Test the MissingHSTSHeader reporter with X-Content-Type-Options header.
// The reporter should not report the issue.
func TestMissingContentTypeOptionsNoIssues(t *testing.T) {
	reporter := reporters.NewMissingContentTypeOptionsReporter()
	if reporter.ErrorType != reporter_errors.ErrorContentTypeOptions {
		t.Errorf("error type is not correct")
	}

	header := &http.Header{}
	header.Set("X-Content-Type-Options", "nosniff")

	// Run the reporter callback with the PageReport.
	reportsIssue := reporter.Callback(&models.PageReport{}, &html.Node{}, header)

	// The reporter should not found any issue.
	if reportsIssue == true {
		t.Errorf("reportsIssue should be false")
	}
}

// Test the MissingHSTSHeader reporter without the X-Content-Type-Options header.
// The reporter should report the issue.
func TestMissingContentTypeOptionsIssues(t *testing.T) {
	reporter := reporters.NewMissingContentTypeOptionsReporter()
	if reporter.ErrorType != reporter_errors.ErrorContentTypeOptions {
		t.Errorf("error type is not correct")
	}

	// Run the reporter callback with the PageReport.
	reportsIssue := reporter.Callback(&models.PageReport{}, &html.Node{}, &http.Header{})

	// The reporter should not found any issue.
	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}
}
