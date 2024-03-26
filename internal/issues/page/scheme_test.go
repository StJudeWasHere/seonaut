package page_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/issues/page"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
)

// Test the HTTPScheme reporter with a PageReport uses the https scheme.
// The reporter should not report the issue.
func TestHTTPSchemeNoIssues(t *testing.T) {
	// Parse an URL with the https scheme.
	pageURL := "https://example.com"
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		t.Errorf("Parse URL error: %v", err)
	}

	// Create a PageReport.
	pageReport := &models.PageReport{
		Crawled:    true,
		URL:        pageURL,
		ParsedURL:  parsedURL,
		StatusCode: 200,
	}

	// Create a new HTTPSchemeReporter.
	reporter := page.NewHTTPSchemeReporter()
	if reporter.ErrorType != errors.ErrorHTTPScheme {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	// Run the reporter callback with the PageReport.
	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	// The reporter should not found any issue.
	if reportsIssue == true {
		t.Errorf("TestHTTPSchemeNoIssues: reportsIssue should be false")
	}
}

// Test the HTTPScheme reporter with a PageReport that uses the http scheme.
// The reporter should report the issue.
func TestHTTPSchemeIssues(t *testing.T) {
	// Parse an URL with the http scheme
	pageURL := "http://example.com"
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		t.Errorf("Parse URL error: %v", err)
	}

	// Create a PageReport.
	pageReport := &models.PageReport{
		Crawled:    true,
		URL:        pageURL,
		ParsedURL:  parsedURL,
		StatusCode: 200,
	}

	// Create a new HTTPSchemeReporter.
	reporter := page.NewHTTPSchemeReporter()
	if reporter.ErrorType != errors.ErrorHTTPScheme {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	// Run the reporter callback with the PageReport.
	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	// The reporter should found an issue.
	if reportsIssue == false {
		t.Errorf("TestHTTPSchemeIssues: reportsIssue should be true")
	}
}
