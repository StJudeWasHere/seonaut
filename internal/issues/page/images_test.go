package page_test

import (
	"net/http"
	"testing"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/issues/page"
	"github.com/stjudewashere/seonaut/internal/models"

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

	reporter := page.NewAltTextReporter()
	if reporter.ErrorType != errors.ErrorImagesWithNoAlt {
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

	reporter := page.NewAltTextReporter()
	if reporter.ErrorType != errors.ErrorImagesWithNoAlt {
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

	reporter := page.NewLongAltTextReporter()
	if reporter.ErrorType != errors.ErrorLongAltText {
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

	reporter := page.NewLongAltTextReporter()
	if reporter.ErrorType != errors.ErrorLongAltText {
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

	reporter := page.NewLargeImageReporter()
	if reporter.ErrorType != errors.ErrorLargeImage {
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

	reporter := page.NewLargeImageReporter()
	if reporter.ErrorType != errors.ErrorLargeImage {
		t.Errorf("error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}
}

// Test the NewNoImageIndex reporter with a pageReport that does not
// have the noimageindex rule in the robots meta tag.
// The reporter should not report the issue.
func TestNoImageIndexReporterNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "text/html",
	}

	reporter := page.NewNoImageIndexReporter()
	if reporter.ErrorType != errors.ErrorNoImageIndex {
		t.Errorf("error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == true {
		t.Errorf("reportsIssue should be false")
	}
}

// Test the NewNoImageIndex reporter with a pageReport that has the noimageindex rule
// in the robots meta tag. The reporter should not report the issue.
func TestImageIndexReporterNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "text/html",
		Robots:    "noimageindex",
	}

	reporter := page.NewNoImageIndexReporter()
	if reporter.ErrorType != errors.ErrorNoImageIndex {
		t.Errorf("error type is not correct")
	}

	reportsIssue := reporter.Callback(pageReport, &html.Node{}, &http.Header{})

	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}
}
