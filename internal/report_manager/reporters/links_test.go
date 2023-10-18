package reporters_test

import (
	"net/url"
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporters"

	"golang.org/x/net/html"
)

// Test the TooManyLinks reporter with a pageReport that does not have too many links.
// The reporter should not report the issue.
func TestTooManyLinksNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	for i := 0; i <= 10; i++ {
		pageReport.Links = append(pageReport.Links, models.Link{})
	}

	reporter := reporters.NewTooManyLinksReporter()
	if reporter.ErrorType != reporter_errors.ErrorTooManyLinks {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

	if reportsIssue == true {
		t.Errorf("TestTooManyLinksNoIssues: reportsIssue should be false")
	}
}

// Test the TooManyLinks reporter with a pageReport that has too many links.
// The reporter should report the issue.
func TestTooManyLinksIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	for i := 0; i <= 110; i++ {
		pageReport.Links = append(pageReport.Links, models.Link{})
	}

	reporter := reporters.NewTooManyLinksReporter()
	if reporter.ErrorType != reporter_errors.ErrorTooManyLinks {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

	if reportsIssue == false {
		t.Errorf("TestTooManyLinksIssues: reportsIssue should be true")
	}
}

// Test the InternalNoFollowLinks reporter with a pageReport that does not have links with
// the nofollow attribute. The reporter should not report the issue.
func TestInternalNoFollowLinksNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	pageReport.Links = append(pageReport.Links, models.Link{})

	reporter := reporters.NewInternalNoFollowLinksReporter()
	if reporter.ErrorType != reporter_errors.ErrorInternalNoFollow {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

	if reportsIssue == true {
		t.Errorf("TestInternalNoFollowLinksNoIssues: reportsIssue should be false")
	}
}

// Test the InternalNoFollowLinks reporter with a pageReport that has links with
// the nofollow attribute. The reporter should report the issue.
func TestInternalNoFollowLinksIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	pageReport.Links = append(pageReport.Links, models.Link{NoFollow: true})

	reporter := reporters.NewInternalNoFollowLinksReporter()
	if reporter.ErrorType != reporter_errors.ErrorInternalNoFollow {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

	if reportsIssue == false {
		t.Errorf("TestInternalNoFollowLinksIssues: reportsIssue should be true")
	}
}

// Test the ExternalLinkWitoutNoFollow reporter with a pageReport that does not have external
// links without the nofollow attribute. The reporter should not report the issue.
func TestExternalLinkWitoutNoFollowNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	pageReport.ExternalLinks = append(pageReport.ExternalLinks, models.Link{NoFollow: true})

	reporter := reporters.NewExternalLinkWitoutNoFollowReporter()
	if reporter.ErrorType != reporter_errors.ErrorExternalWithoutNoFollow {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

	if reportsIssue == true {
		t.Errorf("TestExternalLinkWitoutNoFollowNoIssues: reportsIssue should be false")
	}
}

// Test the ExternalLinkWitoutNoFollow reporter with a pageReport that has external
// links without the nofollow attribute. The reporter should report the issue.
func TestExternalLinkWitoutNoFollowIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	pageReport.ExternalLinks = append(pageReport.ExternalLinks, models.Link{})

	reporter := reporters.NewExternalLinkWitoutNoFollowReporter()
	if reporter.ErrorType != reporter_errors.ErrorExternalWithoutNoFollow {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

	if reportsIssue == false {
		t.Errorf("TestExternalLinkWitoutNoFollowIssues: reportsIssue should be true")
	}
}

// Test the HTTPLinks reporter with a pageReport that does not have links with the http scheme.
// The reporter should not report the issue.
func TestHTTPLinksNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	u := "https://example.com"

	parsedURL, err := url.Parse(u)
	if err != nil {
		t.Errorf("%v", err)
	}

	link := models.Link{
		URL:       u,
		ParsedURL: parsedURL,
	}

	pageReport.Links = append(pageReport.Links, link)

	reporter := reporters.NewHTTPLinksReporter()
	if reporter.ErrorType != reporter_errors.ErrorHTTPLinks {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

	if reportsIssue == true {
		t.Errorf("TestHTTPLinksNoIssues: reportsIssue should be false")
	}
}

// Test the HTTPLinks reporter with a pageReport that has links with the http scheme.
// The reporter should not report the issue.
func TestHTTPLinksIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	u := "http://example.com"

	parsedURL, err := url.Parse(u)
	if err != nil {
		t.Errorf("%v", err)
	}

	link := models.Link{
		URL:       u,
		ParsedURL: parsedURL,
	}

	pageReport.Links = append(pageReport.Links, link)

	reporter := reporters.NewHTTPLinksReporter()
	if reporter.ErrorType != reporter_errors.ErrorHTTPLinks {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

	if reportsIssue == false {
		t.Errorf("TestHTTPLinksIssues: reportsIssue should be true")
	}
}

// Test the Deadend reporter with a pageReport that has links.
// The reporter should not report the issue.
func TestDeadendNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	u := "https://example.com"

	parsedURL, err := url.Parse(u)
	if err != nil {
		t.Errorf("%v", err)
	}

	link := models.Link{
		URL:       u,
		ParsedURL: parsedURL,
	}

	pageReport.Links = append(pageReport.Links, link)

	reporter := reporters.NewDeadendReporter()
	if reporter.ErrorType != reporter_errors.ErrorDeadend {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

	if reportsIssue == true {
		t.Errorf("TestHTTPLinksIssues: reportsIssue should be false")
	}
}

// Test the Deadend reporter with a pageReport that does not have links.
// The reporter should not report the issue.
func TestDeadendIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := reporters.NewDeadendReporter()
	if reporter.ErrorType != reporter_errors.ErrorDeadend {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

	if reportsIssue == false {
		t.Errorf("TestHTTPLinksIssues: reportsIssue should be true")
	}
}
