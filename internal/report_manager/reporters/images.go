package reporters

import (
	"net/http"
	"strings"

	"golang.org/x/net/html"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Returns a report_manager.PageIssueReporter with a callback function to check
// if a page has images with no alt attribute. The callback returns true in case
// the page is text/html and contains images with empty or missing alt attribute.
func NewAltTextReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		for _, i := range pageReport.Images {
			if i.Alt == "" {
				return true
			}
		}

		return false
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorImagesWithNoAlt,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function to check
// if the page report is a large image, in wich case it will return true.
func NewLargeImageReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		return strings.HasPrefix(pageReport.MediaType, "image") && pageReport.Size > 500000
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorLargeImage,
		Callback:  c,
	}
}
