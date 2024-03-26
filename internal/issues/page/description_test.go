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

// Test the EmptyDescription reporter with a pageReport that has description.
// The reporter should not report the issue.
func TestEmptyDescriptionNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:     true,
		MediaType:   "text/html",
		StatusCode:  200,
		Description: "not empty description",
	}

	reporter := page.NewEmptyDescriptionReporter()
	if reporter.ErrorType != errors.ErrorEmptyDescription {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestEmptyDescriptionNoIssues: reportsIssue should be false")
	}
}

// Test the EmptyDescription reporter with a pageReport that doesn't have description.
// The reporter should report the issue.
func TestEmptyDescriptionIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := page.NewEmptyDescriptionReporter()
	if reporter.ErrorType != errors.ErrorEmptyDescription {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestEmptyDescriptionIssues: reportsIssue should be true")
	}
}

// Test the ShortDescription reporter with a pageReport that doesn't have a short description.
// The reporter should not report the issue.
func TestShortDescriptionNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Description: `
			This test should return false if the pageReport description is not short.
			This test should return false if the pageReport description is not short`,
	}

	reporter := page.NewShortDescriptionReporter()
	if reporter.ErrorType != errors.ErrorShortDescription {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestShortDescriptionNoIssues: reportsIssue should be false")
	}
}

// Test the ShortDescription reporter with a pageReport that has a short description.
// The reporter should report the issue.
func TestShortDescriptionIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:     true,
		MediaType:   "text/html",
		StatusCode:  200,
		Description: "This test should return true",
	}

	reporter := page.NewShortDescriptionReporter()
	if reporter.ErrorType != errors.ErrorShortDescription {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestShortDescriptionIssues: reportsIssue should be true")
	}
}

// Test the LongDescription reporter with a pageReport that doesn't have a long description.
// The reporter should not report the issue.
func TestLongDescriptionNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Description: `
			This test should return false if the pageReport description is not short.
			This test should return false if the pageReport description is not short`,
	}

	reporter := page.NewLongDescriptionReporter()
	if reporter.ErrorType != errors.ErrorLongDescription {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestLongDescriptionNoIssues: reportsIssue should be false")
	}
}

// Test the LongDescription reporter with a pageReport that has a long description.
// The reporter should report the issue.
func TestLongDescriptionIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Description: `
			This test should return false if the pageReport description is not short.
			This test should return false if the pageReport description is not short.
			This test should return false if the pageReport description is not short.
			This test should return false if the pageReport description is not short`,
	}

	reporter := page.NewLongDescriptionReporter()
	if reporter.ErrorType != errors.ErrorLongDescription {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestLongDescriptionIssues: reportsIssue should be true")
	}
}

// Test the MultipleDescriptionTags reporter with a page that has one description tag.
// The reporter should not report any issue.
func TestMultipleDescriptionTagsNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := page.NewMultipleDescriptionTagsReporter()
	if reporter.ErrorType != errors.ErrorMultipleDescriptionTags {
		t.Errorf("error type is not correct")
	}

	source := `
	<html>
		<head>
			<meta name="description" content="Test Page Description" />
		</head>
		<body></body>
	</html>
	`

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == true {
		t.Errorf("reportsIssue should be false")
	}
}

// Test the MultipleDescriptionTags reporter with a page that has more than one description tag.
// The reporter should report an issue.
func TestMultipleDescriptionTagsIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := page.NewMultipleDescriptionTagsReporter()
	if reporter.ErrorType != errors.ErrorMultipleDescriptionTags {
		t.Errorf("error type is not correct")
	}

	source := `
	<html>
		<head>
			<meta name="description" content="Test Page Description 1" />
			<meta name="description" content="Test Page Description 2" />
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
