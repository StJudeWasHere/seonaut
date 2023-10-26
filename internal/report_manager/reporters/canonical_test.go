package reporters_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporters"

	"golang.org/x/net/html"
)

// Test the CanonicalMultipleTags reporter with an html source code that has one
// canonical tags. The reporter should not report the issue.
func TestMultipleCanonicalTagsNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	source := `
		<html>
			<head>
				<link rel="canonical" href="https://example.com/home" />
			</head>
		</html>`

	reporter := reporters.NewCanonicalMultipleTagsReporter()
	if reporter.ErrorType != reporter_errors.ErrorMultipleCanonicalTags {
		t.Errorf("CanonicalMultipleTags: error type is not correct")
	}

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("CanonicalMultipleTags: Error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == true {
		t.Errorf("CanonicalMultipleTags: reportsIssue should be false")
	}
}

// Test the CanonicalMultipleTags reporter with an html source code that has multiple
// canonical tags. The reporter should report the issue.
func TestMultipleCanonicalTagsIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	source := `
		<html>
			<head>
				<link rel="canonical" href="https://example.com/home" />
				<link rel="canonical" href="https://example.com/home-2" />
			</head>
		</html>`

	reporter := reporters.NewCanonicalMultipleTagsReporter()
	if reporter.ErrorType != reporter_errors.ErrorMultipleCanonicalTags {
		t.Errorf("MultipleCanonicalTags: error type is not correct")
	}

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("CanonicalMultipleTags: Error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == false {
		t.Errorf("CanonicalMultipleTags: reportsIssue should be true")
	}
}

// Test the CanonicalRelativeURL reporter with an html source code that has a
// canonical tag with an absolute URL.
// The reporter should not report the issue.
func TestCanonicalTagsRelativeURLNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	source := `
		<html>
			<head>
				<link rel="canonical" href="https://example.com/home" />
			</head>
		</html>`

	reporter := reporters.NewCanonicalRelativeURLReporter()
	if reporter.ErrorType != reporter_errors.ErrorRelativeCanonicalURL {
		t.Errorf("CanonicalTagsRelative: error type is not correct")
	}

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("CanonicalTagsRelative: Error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == true {
		t.Errorf("CanonicalTagsRelative: reportsIssue should be false")
	}
}

// Test the CanonicalRelativeURL reporter with an html source code that has a
// canonical tag with a relative URL.
// The reporter should report the issue.
func TestCanonicalTagsRelativeURLIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	source := `
		<html>
			<head>
				<link rel="canonical" href="/home" />
			</head>
		</html>`

	reporter := reporters.NewCanonicalRelativeURLReporter()
	if reporter.ErrorType != reporter_errors.ErrorRelativeCanonicalURL {
		t.Errorf("CanonicalTagsRelative: error type is not correct")
	}

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("CanonicalTagsRelative: Error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == false {
		t.Errorf("CanonicalTagsRelative: reportsIssue should be true")
	}
}

// Test the CanonicalMismatch reporter with an html source code that has the
// same canonical tags in the HTML code and in the HTTP header.
// The reporter should not report the issue.
func TestCanonicalMismatchNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	source := `
		<html>
			<head>
				<link rel="canonical" href="https://example.com/home" />
			</head>
		</html>`

	reporter := reporters.NewCanonicalMismatch()
	if reporter.ErrorType != reporter_errors.ErrorCanonicalMismatch {
		t.Errorf("CanonicalTagsRelative: error type is not correct")
	}

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("CanonicalTagsRelative: Error parsing html source")
	}

	header := &http.Header{}
	header.Set("Link", "<https://example.com/home>; rel=\"canonical\"")

	reportsIssue := reporter.Callback(pageReport, doc, header)

	if reportsIssue == true {
		t.Errorf("CanonicalTagsRelative: reportsIssue should be false")
	}
}

// Test the CanonicalMismatch reporter with an html source code that has different
// canonical tags in the HTML code and in the HTTP header.
// The reporter should report the issue.
func TestCanonicalMismatchIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	source := `
		<html>
			<head>
				<link rel="canonical" href="https://example.com/home" />
			</head>
		</html>`

	reporter := reporters.NewCanonicalMismatch()
	if reporter.ErrorType != reporter_errors.ErrorCanonicalMismatch {
		t.Errorf("CanonicalTagsRelative: error type is not correct")
	}

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("CanonicalTagsRelative: Error parsing html source")
	}

	header := &http.Header{}
	header.Set("Link", "<https://example.com/home-2>; rel=\"canonical\"")

	reportsIssue := reporter.Callback(pageReport, doc, header)

	if reportsIssue == false {
		t.Errorf("CanonicalTagsRelative: reportsIssue should be true")
	}
}
