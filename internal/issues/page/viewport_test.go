package page_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/issues/page"
	"github.com/stjudewashere/seonaut/internal/models"
)

// Test the NewViewportTag reporter with a document that contains a meta viewport tag with content.
// The reporter should not report the issue.
func TestViewportNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	html := strings.NewReader(`
	<html>
		<head>
			<meta name="viewport" content="width=device-width,initial-scale=1">
		</head>
		<body>main text</body>
	</html>`)

	doc, err := htmlquery.Parse(html)
	if err != nil {
		t.Errorf("TestValidHeadingsOrderIssues: error parsing html")
	}

	reporter := page.NewViewportTagReporter()
	if reporter.ErrorType != errors.ErrorMissingViewportTag {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == true {
		t.Errorf("reportsIssue should be false")
	}
}

// Test the NewViewportTag reporter with a document that doesn't contains a meta viewport tag.
// The reporter should report the issue.
func TestViewportIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	html := strings.NewReader(`
	<html>
		<head></head>
		<body>main text</body>
	</html>`)

	doc, err := htmlquery.Parse(html)
	if err != nil {
		t.Errorf("TestValidHeadingsOrderIssues: error parsing html")
	}

	reporter := page.NewViewportTagReporter()
	if reporter.ErrorType != errors.ErrorMissingViewportTag {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}
}

// Test the NewViewportTag reporter with a document that contains an empty meta viewport tag.
// The reporter should report the issue.
func TestEmptyViewportIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	html := strings.NewReader(`
	<html>
		<head>
			<meta name="viewport" content="">
		</head>
		<body>main text</body>
	</html>`)

	doc, err := htmlquery.Parse(html)
	if err != nil {
		t.Errorf("TestValidHeadingsOrderIssues: error parsing html")
	}

	reporter := page.NewViewportTagReporter()
	if reporter.ErrorType != errors.ErrorMissingViewportTag {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}
}
