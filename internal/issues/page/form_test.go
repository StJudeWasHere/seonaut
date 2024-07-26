package page_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/issues/page"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
)

// Test the FormOnHTTP reporter with a pageReport that does not
// have a form on an HTTP URL issue. The reporter should not report the issue.
func TestFormOnHTTPReporterNoIssues(t *testing.T) {
	u := "https://example.com"
	parsedURL, err := url.Parse(u)
	if err != nil {
		t.Errorf("url.Parse: %v", err)
	}

	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Words:      300,
		URL:        u,
		ParsedURL:  parsedURL,
	}

	s := `<html><body><form action=""></form></body</html>`
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		t.Errorf("html.Parse: %v", err)
	}

	reporter := page.NewFormOnHTTPReporter()
	if reporter.ErrorType != errors.ErrorFormOnHTTP {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})
	if reportsIssue == true {
		t.Errorf("TestFormOnHTTPReporterNoIssues: reportsIssue should be false")
	}

	// Test without forms
	doc = &html.Node{}
	reportsIssue = reporter.Callback(pageReport, doc, &http.Header{})
	if reportsIssue == true {
		t.Errorf("TestFormOnHTTPReporterNoIssues empty body: reportsIssue should be false")
	}
}

// Test the FormOnHTTP reporter with a pageReport that does have
// a form on an HTTP URL issue. The reporter should report the issue.
func TestFormOnHTTPReporterIssues(t *testing.T) {
	u := "http://example.com"
	parsedURL, err := url.Parse(u)
	if err != nil {
		t.Errorf("url.Parse: %v", err)
	}

	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Words:      300,
		URL:        u,
		ParsedURL:  parsedURL,
	}

	s := `<html><body><form action=""></form></body</html>`
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		t.Errorf("html.Parse: %v", err)
	}

	reporter := page.NewFormOnHTTPReporter()
	if reporter.ErrorType != errors.ErrorFormOnHTTP {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})
	if reportsIssue == false {
		t.Errorf("TestFormOnHTTPReporterIssues: reportsIssue should be true")
	}
}

// Test the InsecureForm reporter with a pageReport that does not
// have a form with an insecure URL. The reporter should not report the issue.
func TestInsecureFormNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Words:      300,
	}

	s := `<html><body><form action="https://example.com"></form></body</html>`
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		t.Errorf("html.Parse: %v", err)
	}

	reporter := page.NewInsecureFormReporter()

	if reporter.ErrorType != errors.ErrorInsecureForm {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})
	if reportsIssue == true {
		t.Errorf("TestInsecureFormNoIssues: reportsIssue should be false")
	}

	// Test without forms
	doc = &html.Node{}
	reportsIssue = reporter.Callback(pageReport, doc, &http.Header{})
	if reportsIssue == true {
		t.Errorf("TestInsecureFormNoIssues empty body: reportsIssue should be false")
	}
}

// Test the InsecureForm reporter with a pageReport that does have
// a form with an insecure URL. The reporter should report the issue.
func TestInsecureFormIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Words:      300,
	}

	s := `<html><body><form action="http://example.com"></form></body</html>`
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		t.Errorf("html.Parse: %v", err)
	}

	reporter := page.NewInsecureFormReporter()
	if reporter.ErrorType != errors.ErrorInsecureForm {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})
	if reportsIssue == false {
		t.Errorf("TestInsecureFormIssues: reportsIssue should be true")
	}
}
