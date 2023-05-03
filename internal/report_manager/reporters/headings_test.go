package reporters_test

import (
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporters"
)

// Test the NoH1 reporter with a pageReport that has an H1 heading.
// The reporter should not report the issue.
func TestNoH1NoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		H1:         "H1 heading",
	}

	reporter := reporters.NewNoH1Reporter()
	if reporter.ErrorType != reporter_errors.ErrorNoH1 {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport)

	if reportsIssue == true {
		t.Errorf("TestNoH1NoIssues: reportsIssue should be false")
	}
}

// Test the NoH1 reporter with a pageReport that does not have an H1 heading.
// The reporter should report the issue.
func TestNoH1Issues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := reporters.NewNoH1Reporter()
	if reporter.ErrorType != reporter_errors.ErrorNoH1 {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport)

	if reportsIssue == false {
		t.Errorf("TestNoH1Issues: reportsIssue should be true")
	}
}

// Test the ValidHeadingsOrder reporter with a pageReport that has a valid heading order.
// The reporter should not report the issue.
func TestValidHeadingsOrderNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:       true,
		MediaType:     "text/html",
		StatusCode:    200,
		ValidHeadings: true,
	}

	reporter := reporters.NewValidHeadingsOrderReporter()
	if reporter.ErrorType != reporter_errors.ErrorNotValidHeadings {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport)

	if reportsIssue == true {
		t.Errorf("TestValidHeadingsOrderNoIssues: reportsIssue should be false")
	}
}

// Test the ValidHeadingsOrder reporter with a pageReport that doesn't have a valid heading order.
// The reporter should report the issue.
func TestValidHeadingsOrderIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := reporters.NewValidHeadingsOrderReporter()
	if reporter.ErrorType != reporter_errors.ErrorNotValidHeadings {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport)

	if reportsIssue == false {
		t.Errorf("TestValidHeadingsOrderIssues: reportsIssue should be true")
	}
}
