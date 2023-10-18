package reporters_test

import (
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporters"

	"golang.org/x/net/html"
)

// Test the NoIndexable reporter with an indexable pageReport.
// The reporter should not report the issue.
func TestNoIndexableNoIssues(t *testing.T) {
	pageReport := &models.PageReport{}

	reporter := reporters.NewNoIndexableReporter()
	if reporter.ErrorType != reporter_errors.ErrorNoIndexable {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

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

	reporter := reporters.NewNoIndexableReporter()
	if reporter.ErrorType != reporter_errors.ErrorNoIndexable {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

	if reportsIssue == false {
		t.Errorf("TestNoIndexableIssues: reportsIssue should be true")
	}
}

// Test the NoIndexable reporter with PageReport that is not blocked by the robots.txt file.
// The reporter should not report the issue.
func TestBlockedByRobotstxtNoIssues(t *testing.T) {
	pageReport := &models.PageReport{}

	reporter := reporters.NewBlockedByRobotstxtReporter()
	if reporter.ErrorType != reporter_errors.ErrorBlocked {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

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

	reporter := reporters.NewBlockedByRobotstxtReporter()
	if reporter.ErrorType != reporter_errors.ErrorBlocked {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

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

	reporter := reporters.NewNoIndexInSitemapReporter()
	if reporter.ErrorType != reporter_errors.ErrorSitemapNoIndex {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

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

	reporter := reporters.NewNoIndexInSitemapReporter()
	if reporter.ErrorType != reporter_errors.ErrorSitemapNoIndex {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

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

	reporter := reporters.NewSitemapAndBlockedReporter()
	if reporter.ErrorType != reporter_errors.ErrorSitemapBlocked {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

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

	reporter := reporters.NewSitemapAndBlockedReporter()
	if reporter.ErrorType != reporter_errors.ErrorSitemapBlocked {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

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

	reporter := reporters.NewNonCanonicalInSitemapReporter()
	if reporter.ErrorType != reporter_errors.ErrorSitemapNonCanonical {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

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

	reporter := reporters.NewNonCanonicalInSitemapReporter()
	if reporter.ErrorType != reporter_errors.ErrorSitemapNonCanonical {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{})

	if reportsIssue == false {
		t.Errorf("TestNonCanonicalInSitemapIssues: reportsIssue should be true")
	}
}
