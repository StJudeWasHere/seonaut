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
		t.Errorf("TestUnderscoreURL: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestUnderscoreURL: reportsIssue should be true")
	}
}

// Test the SpaceURL reporter with an URL that has not a space character.
// The reporter should not report the issue.
func TestNoSpaceURL(t *testing.T) {
	pageReport := &models.PageReport{
		URL:        "https://example.com/someurl",
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Title:      "not empty description",
	}

	reporter := page.NewSpaceURLReporter()
	if reporter.ErrorType != errors.ErrorSpaceURL {
		t.Errorf("TestNoSpaceURL: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestNoSpaceURL: reportsIssue should be false")
	}
}

// Test the UnderscoreURL reporter with an URL that has a spave character.
// The reporter should report the issue.
func TestSpaceURL(t *testing.T) {
	pageReport := &models.PageReport{
		URL:        "https://example.com/some url",
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Title:      "not empty description",
	}

	reporter := page.NewSpaceURLReporter()
	if reporter.ErrorType != errors.ErrorSpaceURL {
		t.Errorf("TestSpaceURL: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestSpaceURL: reportsIssue should be true")
	}
}

// Test the MultipleSlashes reporter with an URL that has not multiple slashes.
// The reporter should not report the issue.
func TestNoMultipleSlashes(t *testing.T) {
	u := "https://example.com/someurl"
	parsedURL, err := url.Parse(u)
	if err != nil {
		t.Errorf("TestNoMultipleSlashes: url parse error %v", err)
	}

	pageReport := &models.PageReport{
		URL:        u,
		ParsedURL:  parsedURL,
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Title:      "not empty description",
	}

	reporter := page.NewMultipleSlashesReporter()
	if reporter.ErrorType != errors.ErrorMultipleSlashes {
		t.Errorf("TestNoMultipleSlashes: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestNoMultipleSlashes: reportsIssue should be false")
	}
}

// Test the MultipleSlashes reporter with an URL that has multiple slashes.
// The reporter should report the issue.
func TestMultiplSlashes(t *testing.T) {
	u := "https://example.com/someurl//"
	parsedURL, err := url.Parse(u)
	if err != nil {
		t.Errorf("TestMultipleSlashes: url parse error %v", err)
	}

	pageReport := &models.PageReport{
		URL:        u,
		ParsedURL:  parsedURL,
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Title:      "not empty description",
	}

	reporter := page.NewMultipleSlashesReporter()
	if reporter.ErrorType != errors.ErrorMultipleSlashes {
		t.Errorf("TestMultiplSlashes: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestMultiplSlashes: reportsIssue should be true")
	}
}
