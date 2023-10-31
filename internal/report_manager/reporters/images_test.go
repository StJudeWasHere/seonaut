package reporters_test

import (
	"net/http"
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporters"

	"golang.org/x/net/html"
)

// Test the AltText reporter with a pageReport that does not
// have any image without Alt text. The reporter should not report the issue.
func TestAltTextReporterNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "text/html",
	}

	// Add an image with alt text
	pageReport.Images = append(pageReport.Images, models.Image{
		Alt: "Image alt text",
	})

	reporter := reporters.NewAltTextReporter()
	if reporter.ErrorType != reporter_errors.ErrorImagesWithNoAlt {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("TestAltTextReporterNoIssues: reportsIssue should be false")
	}
}

// Test the LittleContent reporter with a pageReport that does
// have a little content issue. The reporter should report the issue.
func TestAltTextReporterIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "text/html",
	}

	// Add an image without alt text
	pageReport.Images = append(pageReport.Images, models.Image{})

	reporter := reporters.NewAltTextReporter()
	if reporter.ErrorType != reporter_errors.ErrorImagesWithNoAlt {
		t.Errorf("TestNoIssues: error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("TestAltTextReporterIssues: reportsIssue should be true")
	}
}

// Test the LongAltText reporter with a pageReport that does not
// have any image with long Alt text. The reporter should not report the issue.
func TestLongAltTextReporterNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "text/html",
	}

	// Add an image with alt text
	pageReport.Images = append(pageReport.Images, models.Image{
		Alt: "Image alt text",
	})

	reporter := reporters.NewLongAltTextReporter()
	if reporter.ErrorType != reporter_errors.ErrorLongAltText {
		t.Errorf("error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("reportsIssue should be false")
	}
}

// Test the LongAltText reporter with a pageReport that does
// have images with long Alt text. The reporter should report the issue.
func TestLongAltTextReporterIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "text/html",
	}

	// Add an image with alt text
	pageReport.Images = append(pageReport.Images, models.Image{
		Alt: "This is a long alt text. This is a long alt text. This is a long alt text. This is a long alt text. This is a long alt text.",
	})

	reporter := reporters.NewLongAltTextReporter()
	if reporter.ErrorType != reporter_errors.ErrorLongAltText {
		t.Errorf("error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}
}

// Test the LargeImage reporter with an image that is not too large.
// The reporter should not report the issue.
func TestLargeImageReporterNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "image/jpeg",
		Size:      300000,
	}

	reporter := reporters.NewLargeImageReporter()
	if reporter.ErrorType != reporter_errors.ErrorLargeImage {
		t.Errorf("error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("reportsIssue should be false")
	}
}

// Test the LargeImage reporter with a large image.
// The reporter should report the issue.
func TestLargeImageReporterIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "image/jpeg",
		Size:      700000,
	}

	reporter := reporters.NewLargeImageReporter()
	if reporter.ErrorType != reporter_errors.ErrorLargeImage {
		t.Errorf("error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}
}
