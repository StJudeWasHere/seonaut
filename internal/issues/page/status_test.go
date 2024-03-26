package page_test

import (
	"net/http"
	"testing"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/issues/page"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
)

// Test the Status30x reporter with a PageReport that has an status code in the 20x range.
// The reporter should not report the issue.
func TestStatus30xNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		StatusCode: 200,
	}

	reporter := page.NewStatus30xReporter()
	if reporter.ErrorType != errors.Error30x {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestStatus30xNoIssues: reportsIssue should be false")
	}
}

// Test the Status30x reporter with a PageReport that has an status code in the 30x range.
// The reporter should report the issue.
func TestStatus30xIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		StatusCode: 301,
	}

	reporter := page.NewStatus30xReporter()
	if reporter.ErrorType != errors.Error30x {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestStatus30xIssues: reportsIssue should be true")
	}
}

// Test the Status40x reporter with a PageReport that has an status code in the 20x range.
// The reporter should not report the issue.
func TestStatus40xNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		StatusCode: 200,
	}

	reporter := page.NewStatus40xReporter()
	if reporter.ErrorType != errors.Error40x {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestStatus40xNoIssues: reportsIssue should be false")
	}
}

// Test the Status40x reporter with a PageReport that has an status code in the 40x range.
// The reporter should report the issue.
func TestStatus40xIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		StatusCode: 401,
	}

	reporter := page.NewStatus40xReporter()
	if reporter.ErrorType != errors.Error40x {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestStatus40xIssues: reportsIssue should be true")
	}
}

// Test the Status50x reporter with a PageReport that has an status code in the 20x range.
// The reporter should not report the issue.
func TestStatus50xNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		StatusCode: 200,
	}

	reporter := page.NewStatus50xReporter()
	if reporter.ErrorType != errors.Error50x {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestStatus50xNoIssues: reportsIssue should be false")
	}
}

// Test the Status50x reporter with a PageReport that has an status code in the 50x range.
// The reporter should report the issue.
func TestStatus50xIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		StatusCode: 501,
	}

	reporter := page.NewStatus50xReporter()
	if reporter.ErrorType != errors.Error50x {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestStatus50xIssues: reportsIssue should be true")
	}
}
