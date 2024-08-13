package page_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/issues/page"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
)

// Test the NoIndexable reporter with an indexable pageReport.
// The reporter should not report the issue.
func TestNoIndexableNoIssues(t *testing.T) {
	pageReport := &models.PageReport{}

	reporter := page.NewNoIndexableReporter()
	if reporter.ErrorType != errors.ErrorNoIndexable {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestNoIndexableNoIssues: reportsIssue should be false")
	}
}

// Test the NoIndexable reporter with a non-indexable pageReport.
// The reporter should report the issue.
func TestNoIndexableIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Noindex: true,
	}

	reporter := page.NewNoIndexableReporter()
	if reporter.ErrorType != errors.ErrorNoIndexable {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestNoIndexableIssues: reportsIssue should be true")
	}
}

// Test the NoIndexable reporter with PageReport that is not blocked by the robots.txt file.
// The reporter should not report the issue.
func TestBlockedByRobotstxtNoIssues(t *testing.T) {
	pageReport := &models.PageReport{}

	reporter := page.NewBlockedByRobotstxtReporter()
	if reporter.ErrorType != errors.ErrorBlocked {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestBlockedByRobotstxtNoIssues: reportsIssue should be false")
	}
}

// Test the NoIndexable reporter with PageReport that is blocked by the robots.txt file.
// The reporter should report the issue.
func TestBlockedByRobotstxtIssues(t *testing.T) {
	pageReport := &models.PageReport{
		BlockedByRobotstxt: true,
	}

	reporter := page.NewBlockedByRobotstxtReporter()
	if reporter.ErrorType != errors.ErrorBlocked {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestBlockedByRobotstxtIssues: reportsIssue should be true")
	}
}

// Test the NoIndexable reporter with PageReport that is included in the sitemap and it is indexable.
// The reporter should not report the issue.
func TestNoIndexInSitemapNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		InSitemap: true,
	}

	reporter := page.NewNoIndexInSitemapReporter()
	if reporter.ErrorType != errors.ErrorSitemapNoIndex {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestNoIndexInSitemapNoIssues: reportsIssue should be false")
	}
}

// Test the NoIndexable reporter with PageReport that is included in the sitemap and it is not indexable.
// The reporter should report the issue.
func TestNoIndexInSitemapIssues(t *testing.T) {
	pageReport := &models.PageReport{
		InSitemap: true,
		Noindex:   true,
	}

	reporter := page.NewNoIndexInSitemapReporter()
	if reporter.ErrorType != errors.ErrorSitemapNoIndex {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestNoIndexInSitemapIssues: reportsIssue should be true")
	}
}

// Test the SitemapAndBlocked reporter with PageReport that is included in the sitemap and it is not
// blocked by the robots.txt file. The reporter should not report the issue.
func TestSitemapAndBlockedNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		InSitemap: true,
	}

	reporter := page.NewSitemapAndBlockedReporter()
	if reporter.ErrorType != errors.ErrorSitemapBlocked {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestSitemapAndBlockedNoIssues: reportsIssue should be false")
	}
}

// Test the SitemapAndBlocked reporter with PageReport that is included in the sitemap and it is
// blocked by the robots.txt file. The reporter should report the issue.
func TestSitemapAndBlockedIssues(t *testing.T) {
	pageReport := &models.PageReport{
		InSitemap:          true,
		BlockedByRobotstxt: true,
	}

	reporter := page.NewSitemapAndBlockedReporter()
	if reporter.ErrorType != errors.ErrorSitemapBlocked {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestSitemapAndBlockedIssues: reportsIssue should be true")
	}
}

// Test the NonCanonicalInSitemap reporter with PageReport that is included in the sitemap and it is canonical.
// The reporter should not report the issue.
func TestNonCanonicalInSitemapNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		InSitemap: true,
		MediaType: "text/html",
	}

	reporter := page.NewNonCanonicalInSitemapReporter()
	if reporter.ErrorType != errors.ErrorSitemapNonCanonical {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestNonCanonicalInSitemapNoIssues: reportsIssue should be false")
	}
}

// Test the NonCanonicalInSitemap reporter with PageReport that is included in the sitemap and it is not
// canonical. The reporter should report the issue.
func TestNonCanonicalInSitemapIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		InSitemap: true,
		MediaType: "text/html",
		URL:       "https://example.com/non-canonical",
		Canonical: "https://example.com/canoical",
	}

	reporter := page.NewNonCanonicalInSitemapReporter()
	if reporter.ErrorType != errors.ErrorSitemapNonCanonical {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestNonCanonicalInSitemapIssues: reportsIssue should be true")
	}
}

// Test the NewMetasInBodyReporter with PageReport that does not have metas in the document's body.
// The reporter should not report the issue.
func TestMetasInBodyNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "text/html",
	}

	reporter := page.NewMetasInBodyReporter()
	if reporter.ErrorType != errors.ErrorMetasInBody {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	s := `<html><body></body</html>`
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		t.Errorf("html.Parse: %v", err)
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestMetasInBodyNoIssues: reportsIssue should be false")
	}
}

// Test the NewMetasInBodyReporter with PageReport that has metas in the document's body.
// The reporter should not report the issue.
func TestMetasInBodyIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "text/html",
	}

	reporter := page.NewMetasInBodyReporter()
	if reporter.ErrorType != errors.ErrorMetasInBody {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	s := `<html><body><meta name="robots" content="noindex, nofollow" /></body</html>`
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		t.Errorf("html.Parse: %v", err)
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestMetasInBodyIssues: reportsIssue should be true")
	}
}

// Test the NewNosnippetReporter with PageReport that does not have the nosnippet directive.
// The reporter should not report the issue.
func TestNosnippetNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "text/html",
	}

	reporter := page.NewNosnippetReporter()
	if reporter.ErrorType != errors.ErrorNosnippet {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})
	if reportsIssue == true {
		t.Errorf("TestNosnippetNoIssues: reportsIssue should be false")
	}
}

// Test the NewNosnippetReporter with PageReport that has the nosnippet directive.
// The reporter should report the issue.
func TestNosnippetIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "text/html",
		Robots:    "nosnippet",
	}

	reporter := page.NewNosnippetReporter()
	if reporter.ErrorType != errors.ErrorNosnippet {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})
	if reportsIssue == false {
		t.Errorf("TestNosnippetIssues nosnippet: reportsIssue should be true")
	}

	pageReport.Robots = "max-snippet:0"
	reportsIssue = reporter.Callback(pageReport, &html.Node{}, &http.Header{})
	if reportsIssue == false {
		t.Errorf("TestNosnippetIssues max-snippet: reportsIssue should be true")
	}
}
