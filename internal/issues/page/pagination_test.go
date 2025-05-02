package page_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/issues/page"
	"github.com/stjudewashere/seonaut/internal/models"
	"golang.org/x/net/html"
)

// Test the NewPaginationReporter with a document that contains the prev and next attributes
// as well as the actual links. The reporter should not report any issue.
func TestPaginationNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	var err error
	pageReport.ParsedURL, err = url.Parse("https://example.com/")
	if err != nil {
		t.Errorf("Error parsing url %v", err)
	}

	// The reporter checks the links in the pageReport Links slice to avoid
	// extracting all the links again.
	pageReport.Links = append(pageReport.Links, models.Link{URL: "https://example.com/page/1"})
	pageReport.Links = append(pageReport.Links, models.Link{URL: "https://example.com/page/3"})

	source := `
	<html>
		<head>
			<link rel="prev" href="/page/1">
			<link rel="next" href="/page/3">
		</head>
		<body></body>
	</html>
	`

	reporter := page.NewPaginationReporter()
	if reporter.ErrorType != errors.ErrorPaginationLink {
		t.Errorf("error type is not correct")
	}

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("Error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == true {
		t.Errorf("reportsIssue should be false")
	}
}

// Test the NewPaginationReporter with a document that contains the prev and next attributes
// but does not contain the actual links. The reporter should report the issue.
func TestPaginationIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	var err error
	pageReport.ParsedURL, err = url.Parse("https://example.com/")
	if err != nil {
		t.Errorf("Error parsing url %v", err)
	}

	reporter := page.NewPaginationReporter()
	if reporter.ErrorType != errors.ErrorPaginationLink {
		t.Errorf("error type is not correct")
	}

	source := `
	<html>
		<head>
			<link rel="prev" href="/page/1">
			<link rel="next" href="/page/3">
		</head>
		<body></body>
	</html>
	`

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("Error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}
}
