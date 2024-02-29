package reporters_test

import (
	"net/http"
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

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

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

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

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

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

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

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

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

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

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

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

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

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

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

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

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

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

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

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestHTTPLinksIssues: reportsIssue should be true")
	}
}

// Test the NewExternalLinkRedirect reporter with a pageReport that has external
// links without redirect issues. The reporter should not report the issue.
func TestExternalLinkRedirectNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	pageReport.ExternalLinks = append(pageReport.ExternalLinks, models.Link{})
	pageReport.ExternalLinks = append(pageReport.ExternalLinks, models.Link{
		StatusCode: 200,
	})

	reporter := reporters.NewExternalLinkRedirect()
	if reporter.ErrorType != reporter_errors.ErrorExternalLinkRedirect {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestExternalLinkRedirectNoIssues: reportsIssue should be false")
	}
}

// Test the NewExternalLinkRedirect reporter with a pageReport that has external
// links with redirect issues. The reporter should report the issue.
func TestExternalLinkRedirectIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	pageReport.ExternalLinks = append(pageReport.ExternalLinks, models.Link{})
	pageReport.ExternalLinks = append(pageReport.ExternalLinks, models.Link{
		StatusCode: 301,
	})

	reporter := reporters.NewExternalLinkRedirect()
	if reporter.ErrorType != reporter_errors.ErrorExternalLinkRedirect {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestExternalLinkRedirectIssues: reportsIssue should be true")
	}
}

// Test the NewExternalLinkBroken reporter with a pageReport that has valid
// external links. The reporter should not report the issue.
func TestExternalLinkBrokenNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	pageReport.ExternalLinks = append(pageReport.ExternalLinks, models.Link{})
	pageReport.ExternalLinks = append(pageReport.ExternalLinks, models.Link{
		StatusCode: 200,
	})

	reporter := reporters.NewExternalLinkBroken()
	if reporter.ErrorType != reporter_errors.ErrorExternalLinkBroken {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestExternalLinkBrokenNoIssues: reportsIssue should be false")
	}
}

// Test the NewExternalLinkBroken reporter with a pageReport that has broken
// external links. The reporter should report the issue.
func TestExternalLinkBrokenIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	pageReport.ExternalLinks = append(pageReport.ExternalLinks, models.Link{})
	pageReport.ExternalLinks = append(pageReport.ExternalLinks, models.Link{
		StatusCode: 400,
	})

	reporter := reporters.NewExternalLinkBroken()
	if reporter.ErrorType != reporter_errors.ErrorExternalLinkBroken {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestExternalLinkBrokenIssues: reportsIssue should be true")
	}
}
