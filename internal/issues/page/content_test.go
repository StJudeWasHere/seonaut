package page_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/antchfx/htmlquery"
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

// Test NewIncorrectMediaType with URLs that have correct media types.
// It should not report any issue.
func TestIncorrectMediaTypeNoIssues(t *testing.T) {
	u := "https://example.com/no-issues"
	parsedURL, err := url.Parse(u)
	if err != nil {
		t.Errorf("error parsing URL: %v", err)
	}

	pageReport := &models.PageReport{
		MediaType: "text/html",
		URL:       u,
		ParsedURL: parsedURL,
	}

	reporter := page.NewIncorrectMediaTypeReporter()
	if reporter.ErrorType != errors.ErrorIncorrectMediaType {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})
	if reportsIssue == true {
		t.Errorf("reportsIssue should be false")
	}

	// Test javascript extension.
	u = "https://example.com/script.js"
	parsedURL, err = url.Parse(u)
	if err != nil {
		t.Errorf("error parsing URL: %v", err)
	}

	pageReport = &models.PageReport{
		MediaType: "application/javascript",
		URL:       u,
		ParsedURL: parsedURL,
	}

	reportsIssue = reporter.Callback(pageReport, &html.Node{}, &http.Header{})
	if reportsIssue == true {
		t.Errorf("reportsIssue should be false")
	}
}

// Test NewIncorrectMediaType with URLs that have incorrect media types.
// It should report the issues.
func TestIncorrectMediaTypeIssues(t *testing.T) {
	pageReport := &models.PageReport{} // Test missing media type

	reporter := page.NewIncorrectMediaTypeReporter()
	if reporter.ErrorType != errors.ErrorIncorrectMediaType {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})
	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}

	// Test media type that doesn't match the file extension.
	u := "https://example.com/issues.pdf"
	parsedURL, err := url.Parse(u)
	if err != nil {
		t.Errorf("error parsing URL: %v", err)
	}

	pageReport = &models.PageReport{
		MediaType: "text/html",
		URL:       u,
		ParsedURL: parsedURL,
	}

	reportsIssue = reporter.Callback(pageReport, &html.Node{}, &http.Header{})
	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}
}

func TestDuplicatedIdIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := page.NewDuplicatedIdReporter()
	if reporter.ErrorType != errors.ErrorDuplicatedId {
		t.Errorf("error type is not correct")
	}

	html := strings.NewReader(`
		<html>
			<body>
				<div id="header">Header 1</div>
				<div id="header2">Header 2</div>
				<span id="header">Header 3</div>
			</body>
		</html>`)

	doc, err := htmlquery.Parse(html)
	if err != nil {
		t.Errorf("error parsing html")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})
	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}
}

func TestDuplicatedIdNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := page.NewDuplicatedIdReporter()
	if reporter.ErrorType != errors.ErrorDuplicatedId {
		t.Errorf("error type is not correct")
	}

	html := strings.NewReader(`
		<html>
			<body>
				<div id="header">Header 1</div>
				<div id="header2">Header 2</div>
				<span id="header3">Header 3</div>
			</body>
		</html>`)

	doc, err := htmlquery.Parse(html)
	if err != nil {
		t.Errorf("error parsing html")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})
	if reportsIssue == true {
		t.Errorf("reportsIssue should be false")
	}
}

// Test the DOMSize reporter with a document with a node count below the specified limit.
// The reporter should not report any issue.
func TestDOMSizeNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := page.NewDOMSizeReporter(1500)
	if reporter.ErrorType != errors.ErrorDOMSize {
		t.Errorf("error type is not correct")
	}

	html := strings.NewReader(`
		<html>
			<body>
				<div>DOM Size Test</div>
			</body>
		</html>`)

	doc, err := htmlquery.Parse(html)
	if err != nil {
		t.Errorf("error parsing html")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})
	if reportsIssue == true {
		t.Errorf("reportsIssue should be false")
	}
}

// Test the DOMSize reporter with a document with a node count higher than the specified size.
// The reporter should report an issue.
func TestDOMSizeIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := page.NewDOMSizeReporter(3)
	if reporter.ErrorType != errors.ErrorDOMSize {
		t.Errorf("error type is not correct")
	}

	html := strings.NewReader(`
		<html>
			<body>
				<div>DOM Size Test</div>
				<div>DOM Size Test</div>
				<div>DOM Size Test</div>
				<div>DOM Size Test</div>
				<div>DOM Size Test</div>
			</body>
		</html>`)

	doc, err := htmlquery.Parse(html)
	if err != nil {
		t.Errorf("error parsing html")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})
	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}
}
