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

// Test the EmptyTitle reporter with a pageReport that has description.
// The reporter should not report the issue.
func TestEmptyTitleNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Title:      "not empty description",
	}

	reporter := page.NewEmptyTitleReporter()
	if reporter.ErrorType != errors.ErrorEmptyTitle {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestEmptyTitleNoIssues: reportsIssue should be false")
	}
}

// Test the EmptyTitle reporter with a pageReport that doesn't have description.
// The reporter should report the issue.
func TestEmptyTitleIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := page.NewEmptyTitleReporter()
	if reporter.ErrorType != errors.ErrorEmptyTitle {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestEmptyTitleIssues: reportsIssue should be true")
	}
}

// Test the ShortTitle reporter with a pageReport that doesn't have a short description.
// The reporter should not report the issue.
func TestShortTitleNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Title: `
			This test should return false if the pageReport description is not short.
			This test should return false if the pageReport description is not short`,
	}

	reporter := page.NewShortTitleReporter()
	if reporter.ErrorType != errors.ErrorShortTitle {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestShortTitleNoIssues: reportsIssue should be false")
	}
}

// Test the ShortTitle reporter with a pageReport that has a short description.
// The reporter should report the issue.
func TestShortTitleIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Title:      "Short title",
	}

	reporter := page.NewShortTitleReporter()
	if reporter.ErrorType != errors.ErrorShortTitle {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestShortTitleIssues: reportsIssue should be true")
	}
}

// Test the LongTitle reporter with a pageReport that doesn't have a long description.
// The reporter should not report the issue.
func TestLongTitleNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Title:      "This test should return false",
	}

	reporter := page.NewLongTitleReporter()
	if reporter.ErrorType != errors.ErrorLongTitle {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestLongTitleNoIssues: reportsIssue should be false")
	}
}

// Test the LongTitle reporter with a pageReport that has a long description.
// The reporter should report the issue.
func TestLongTitleIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Title: `
			This test should return false if the pageReport description is not short.
			This test should return false if the pageReport description is not short.
			This test should return false if the pageReport description is not short.
			This test should return false if the pageReport description is not short`,
	}

	reporter := page.NewLongTitleReporter()
	if reporter.ErrorType != errors.ErrorLongTitle {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestLongTitleIssues: reportsIssue should be true")
	}
}

// Test the MultipleTitleTags reporter with a page that has only one title tag.
// The reporter should not report any issue.
func TestMultipleTitleTagsNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := page.NewMultipleTitleTagsReporter()
	if reporter.ErrorType != errors.ErrorMultipleTitleTags {
		t.Errorf("error type is not correct")
	}

	source := `
	<html>
		<head>
			<title>Title</title>
		</head>
		<body></body>
	</html>
	`

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("Error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == true {
		t.Errorf("reportsIssue should be false")
	}
}

// Test the MultipleTitleTags reporter with a page that has more than one title tag.
// The reporter should report an issue.
func TestMultipleTitleTagsIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := page.NewMultipleTitleTagsReporter()
	if reporter.ErrorType != errors.ErrorMultipleTitleTags {
		t.Errorf("error type is not correct")
	}

	source := `
	<html>
		<head>
			<title>Title 1</title>
			<title>Title 1</title>
		</head>
		<body></body>
	</html>
	`

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}
}
