package page_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/issues/page"
	"github.com/stjudewashere/seonaut/internal/models"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
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

	reporter := page.NewNoH1Reporter()
	if reporter.ErrorType != errors.ErrorNoH1 {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

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

	reporter := page.NewNoH1Reporter()
	if reporter.ErrorType != errors.ErrorNoH1 {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestNoH1Issues: reportsIssue should be true")
	}
}

// Test the ValidHeadingsOrder reporter with a pageReport that has a valid heading order.
// The reporter should not report the issue.
func TestValidHeadingsOrderNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := page.NewValidHeadingsOrderReporter()
	if reporter.ErrorType != errors.ErrorNotValidHeadings {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	html := strings.NewReader(`
		<html>
			<body>
				<h1>Heder 1</h1>
				<h2>Header 2</h2>
				<h3>Header 3</h3>
			</body>
		</html>`)

	doc, err := htmlquery.Parse(html)
	if err != nil {
		t.Errorf("TestValidHeadingsOrderNoIssues: error parsing html")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

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

	reporter := page.NewValidHeadingsOrderReporter()
	if reporter.ErrorType != errors.ErrorNotValidHeadings {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	html := strings.NewReader(`
		<html>
			<body>
				<h1>Header 1</h1>
				<h3>Header 3</h3>
				<h2>Header 2</h2>
				<h4>Header 4</h4>
			</body>
		</html>`)

	doc, err := htmlquery.Parse(html)
	if err != nil {
		t.Errorf("TestValidHeadingsOrderIssues: error parsing html")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestValidHeadingsOrderIssues: reportsIssue should be true")
	}
}
