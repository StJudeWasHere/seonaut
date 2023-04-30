package reporters_test

import (
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporters"
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

	reporter := reporters.NewEmptyDescriptionReporter()
	if reporter.ErrorType != reporters.ErrorEmptyDescription {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport)

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

	reporter := reporters.NewEmptyDescriptionReporter()
	if reporter.ErrorType != reporters.ErrorEmptyDescription {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport)

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

	reporter := reporters.NewShortDescriptionReporter()
	if reporter.ErrorType != reporters.ErrorShortDescription {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport)

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

	reporter := reporters.NewShortDescriptionReporter()
	if reporter.ErrorType != reporters.ErrorShortDescription {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport)

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

	reporter := reporters.NewLongDescriptionReporter()
	if reporter.ErrorType != reporters.ErrorLongDescription {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport)

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

	reporter := reporters.NewLongDescriptionReporter()
	if reporter.ErrorType != reporters.ErrorLongDescription {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport)

	if reportsIssue == false {
		t.Errorf("TestLongDescriptionIssues: reportsIssue should be true")
	}
}
