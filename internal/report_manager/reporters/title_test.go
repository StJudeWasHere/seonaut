package reporters_test

import (
	"net/http"
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporters"

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

	reporter := reporters.NewEmptyTitleReporter()
	if reporter.ErrorType != reporter_errors.ErrorEmptyTitle {
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

	reporter := reporters.NewEmptyTitleReporter()
	if reporter.ErrorType != reporter_errors.ErrorEmptyTitle {
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

	reporter := reporters.NewShortTitleReporter()
	if reporter.ErrorType != reporter_errors.ErrorShortTitle {
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

	reporter := reporters.NewShortTitleReporter()
	if reporter.ErrorType != reporter_errors.ErrorShortTitle {
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

	reporter := reporters.NewLongTitleReporter()
	if reporter.ErrorType != reporter_errors.ErrorLongTitle {
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

	reporter := reporters.NewLongTitleReporter()
	if reporter.ErrorType != reporter_errors.ErrorLongTitle {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestLongTitleIssues: reportsIssue should be true")
	}
}
