package page_test

import (
	"net/http"
	"testing"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/issues/page"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
)

// Test the TestSlowTTFB reporter with a fast TTFB.
// The reporter should not report the issue.
func TestNoSlowTTFB(t *testing.T) {
	pageReport := &models.PageReport{
		URL:  "https://example.com/some-url",
		TTFB: 100,
	}

	reporter := page.NewSlowTTFBReporter()
	if reporter.ErrorType != errors.ErrorSlowTTFB {
		t.Errorf("TestNoSlowTTFB: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestNoSlowTTFB: reportsIssue should be false")
	}
}

// Test the TestSlowTTFB reporter with a slow TTFB.
// The reporter should report the issue.
func TestSlowTTFB(t *testing.T) {
	pageReport := &models.PageReport{
		URL:  "https://example.com/some-url",
		TTFB: 1000,
	}

	reporter := page.NewSlowTTFBReporter()
	if reporter.ErrorType != errors.ErrorSlowTTFB {
		t.Errorf("TestNoSlowTTFB: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestNoSlowTTFB: reportsIssue should be true")
	}
}
