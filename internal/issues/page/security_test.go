package page_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/issues/page"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
)

// Test the MissingHSTSHeader reporter with a PageReport with HSTS header.
// The reporter should report the issue.
func TestMissingHSTSHeaderIssues(t *testing.T) {
	reporter := page.NewMissingHSTSHeaderReporter()
	if reporter.ErrorType != errors.ErrorMissingHSTSHeader {
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
	reporter := page.NewMissingHSTSHeaderReporter()
	if reporter.ErrorType != errors.ErrorMissingHSTSHeader {
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

	reporter := page.NewMissingCSPReporter()
	if reporter.ErrorType != errors.ErrorMissingCSP {
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

	reporter := page.NewMissingCSPReporter()
	if reporter.ErrorType != errors.ErrorMissingCSP {
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
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := page.NewMissingContentTypeOptionsReporter()
	if reporter.ErrorType != errors.ErrorContentTypeOptions {
		t.Errorf("error type is not correct")
	}

	header := &http.Header{}
	header.Set("X-Content-Type-Options", "nosniff")

	// Run the reporter callback with the PageReport.
	reportsIssue := reporter.Callback(pageReport, &html.Node{}, header)

	// The reporter should not found any issue.
	if reportsIssue == true {
		t.Errorf("reportsIssue should be false")
	}
}

// Test the MissingHSTSHeader reporter without the X-Content-Type-Options header.
// The reporter should report the issue.
func TestMissingContentTypeOptionsIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := page.NewMissingContentTypeOptionsReporter()
	if reporter.ErrorType != errors.ErrorContentTypeOptions {
		t.Errorf("error type is not correct")
	}

	// Run the reporter callback with the PageReport.
	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	// The reporter should not found any issue.
	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}
}
