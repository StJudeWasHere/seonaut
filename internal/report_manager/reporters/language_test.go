package reporters_test

import (
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporters"
)

// Test the InvalidLang reporter with a PageReport that has a valid language attribute.
// The reporter should not report the issue.
func TestInvalidLangNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "text/html",
		ValidLang: true,
	}

	reporter := reporters.NewInvalidLangReporter()
	if reporter.ErrorType != reporter_errors.ErrorInvalidLanguage {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport)

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
	}

	reporter := reporters.NewInvalidLangReporter()
	if reporter.ErrorType != reporter_errors.ErrorInvalidLanguage {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport)

	if reportsIssue == false {
		t.Errorf("TestInvalidLangIssues: reportsIssue should be true")
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

	reporter := reporters.NewMissingLangReporter()
	if reporter.ErrorType != reporter_errors.ErrorNoLang {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport)

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

	reporter := reporters.NewMissingLangReporter()
	if reporter.ErrorType != reporter_errors.ErrorNoLang {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport)

	if reportsIssue == false {
		t.Errorf("TestMissingLangIssues: reportsIssue should be true")
	}
}
