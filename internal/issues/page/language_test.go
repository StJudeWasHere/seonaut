package page_test

import (
	"net/http"
	"testing"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/issues/page"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
)

// Test the InvalidLang reporter with a PageReport that has a valid language attribute.
// The reporter should not report the issue.
func TestInvalidLangNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "text/html",
		Lang:      "en",
	}

	reporter := page.NewInvalidLangReporter()
	if reporter.ErrorType != errors.ErrorInvalidLanguage {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestInvalidLangNoIssues: reportsIssue should be false")
	}
}

// Test the InvalidLang reporter with a PageReport that has an invalid language attribute.
// The reporter should report the issue.
func TestInvalidLangIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "text/html",
		Lang:      "InvalidLangCode",
	}

	reporter := page.NewInvalidLangReporter()
	if reporter.ErrorType != errors.ErrorInvalidLanguage {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestInvalidLangIssues: reportsIssue should be true")
	}
}

// Test the InvalidLang reporter with a PageReport that has a redirect status code.
// The reporter should not report the issue.
func TestInvalidLang30xNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		Lang:       "InvalidLangCode",
		StatusCode: 301,
	}

	reporter := page.NewInvalidLangReporter()
	if reporter.ErrorType != errors.ErrorInvalidLanguage {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestInvalidLangIssues: reportsIssue should be false")
	}
}

// Test the MissingLang reporter with a PageReport that has a language attribute.
// The reporter should not report the issue.
func TestMissingLangNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "text/html",
		Lang:      "en",
	}

	reporter := page.NewMissingLangReporter()
	if reporter.ErrorType != errors.ErrorNoLang {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestMissingLangNoIssues: reportsIssue should be false")
	}
}

// Test the MissingLang reporter with a PageReport that has a redirect status code.
// The reporter should not report the issue.
func TestMissingLang30xNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 301,
	}

	reporter := page.NewMissingLangReporter()
	if reporter.ErrorType != errors.ErrorNoLang {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestMissingLangNoIssues: reportsIssue should be false")
	}
}

// Test the MissingLang reporter with a PageReport that has an empty language attribute.
// The reporter should report the issue.
func TestMissingLangIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "text/html",
	}

	reporter := page.NewMissingLangReporter()
	if reporter.ErrorType != errors.ErrorNoLang {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestMissingLangIssues: reportsIssue should be true")
	}
}
