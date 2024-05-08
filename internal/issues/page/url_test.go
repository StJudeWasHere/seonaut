package page_test

import (
	"net/http"
	"testing"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/issues/page"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
)

// Test the UnderscoreURL reporter with an URL that has not an _ character.
// The reporter should not report the issue.
func TestNoUnderscoreURL(t *testing.T) {
	pageReport := &models.PageReport{
		URL:        "https://example.com/some-url",
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Title:      "not empty description",
	}

	reporter := page.NewUnderscoreURLReporter()
	if reporter.ErrorType != errors.ErrorUnderscoreURL {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestUnderscoreURL: reportsIssue should be false")
	}
}

// Test the UnderscoreURL reporter with an URL that has an _ character.
// The reporter should report the issue.
func TestUnderscoreURL(t *testing.T) {
	pageReport := &models.PageReport{
		URL:        "https://example.com/some_url",
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Title:      "not empty description",
	}

	reporter := page.NewUnderscoreURLReporter()
	if reporter.ErrorType != errors.ErrorUnderscoreURL {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestUnderscoreURL: reportsIssue should be true")
	}
}
