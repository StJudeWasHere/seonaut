package reporters_test

import (
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporters"
)

// Test the LittleContent reporter with a pageReport that does not
// have a little content issue. The reporter should not report the issue.
func TestLittelContentNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Words:      300,
	}

	reporter := reporters.NewLittleContentReporter()
	if reporter.ErrorType != reporters.ErrorLittleContent {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport)

	if reportsIssue == true {
		t.Errorf("TestLittelContentNoIssues: reportsIssue should be false")
	}
}

// Test the LittleContent reporter with a pageReport that does
// have a little content issue. The reporter should report the issue.
func TestLittleContentIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Words:      30,
	}

	reporter := reporters.NewLittleContentReporter()
	if reporter.ErrorType != reporters.ErrorLittleContent {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport)

	if reportsIssue == false {
		t.Errorf("TestLittleContentIssues: reportsIssue should be true")
	}
}
