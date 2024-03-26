package page_test

import (
	"net/http"
	"testing"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/issues/page"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
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

	reporter := page.NewLittleContentReporter()
	if reporter.ErrorType != errors.ErrorLittleContent {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

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

	reporter := page.NewLittleContentReporter()
	if reporter.ErrorType != errors.ErrorLittleContent {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestLittleContentIssues: reportsIssue should be true")
	}
}
