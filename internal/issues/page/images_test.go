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
func TestImageIndexReporterIssues(t *testing.T) {
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

// Test the NewNoImageIndex reporter with a pageReport that has all the img elements in
// the pictures. The reporter should not report the issue.
func TestMissingImgTagInPictureReporterNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "text/html",
	}

	reporter := page.NewMissingImgTagInPictureReporter()
	if reporter.ErrorType != errors.ErrorMissingImgElement {
		t.Errorf("error type is not correct")
	}

	source := `
	<html>
		<body>
			<picture>
				<source srcset="/media/img-240-200.jpg" media="(orientation: portrait)" />
				<img src="/media/media/img-298-332.jpg" alt="" />
			</picture>
		</body>
	</html>
`

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == true {
		t.Errorf("reportsIssue should be false")
	}
}

// Test the NewNoImageIndex reporter with a pageReport that hasn't the img elements in
// the pictures. The reporter should report the issue.
func TestMissingImgTagInPictureReporterIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "text/html",
	}

	reporter := page.NewMissingImgTagInPictureReporter()
	if reporter.ErrorType != errors.ErrorMissingImgElement {
		t.Errorf("error type is not correct")
	}

	source := `
	<html>
		<body>
			<picture>
				<source srcset="/media/img-240-200.jpg" media="(orientation: portrait)" />
			</picture>
		</body>
	</html>
`

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}
}

// Test the NewImgWithoutSizeReporter reporter with a pageReport that the img elements
// with size attributes. The reporter should not report the issue.
func TestImgWithoutSizeReporterNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "text/html",
	}

	reporter := page.NewImgWithoutSizeReporter()
	if reporter.ErrorType != errors.ErrorImgWithoutSize {
		t.Errorf("error type is not correct")
	}

	source := `
	<html>
		<body>
			<img src="example.jpg" width="80vw" height="100%">
			<img src="example-2.jpg" width="400" height="400">
		</body>
	</html>
`

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == true {
		t.Errorf("reportsIssue should be false")
	}
}

// Test the NewImgWithoutSizeReporter reporter with a pageReport that the img elements
// without size attributes. The reporter should report the issue.
func TestImgWithoutSizeReporterIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:   true,
		MediaType: "text/html",
	}

	reporter := page.NewImgWithoutSizeReporter()
	if reporter.ErrorType != errors.ErrorImgWithoutSize {
		t.Errorf("error type is not correct")
	}

	source := `
	<html>
		<body>
			<img src="example.jpg">
		</body>
	</html>
`

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}

	// Test img only with the height attribute.
	source = `
	<html>
		<body>
			<img src="example.jpg" height="200">
		</body>
	</html>
`

	doc, err = html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("error parsing html source")
	}

	reportsIssue = reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}

	// Test img only with the width attribute.
	source = `
		<html>
			<body>
				<img src="example.jpg" width="200">
			</body>
		</html>
	`

	doc, err = html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("error parsing html source")
	}

	reportsIssue = reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == false {
		t.Errorf("reportsIssue should be true")
	}
}
