package page

import (
	"net/http"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
)

// Returns a report_manager.PageIssueReporter with a callback function to check
// if a page has images with no alt attribute. The callback returns true in case
// the page is text/html and contains images with empty or missing alt attribute.
func NewAltTextReporter() *models.PageIssueReporter {
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

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorImagesWithNoAlt,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function to check
// if a page has images with a long alt attribute. The callback returns true in case
// the page is text/html and contains images with long alt attribute.
func NewLongAltTextReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		for _, i := range pageReport.Images {
			if len([]rune(i.Alt)) > 100 {
				return true
			}
		}

		return false
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorLongAltText,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function to check
// if the page report is a large image, in wich case it will return true.
func NewLargeImageReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		return strings.HasPrefix(pageReport.MediaType, "image") && pageReport.Size > 500000
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorLargeImage,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function to check
// if a page has the noimageindex rule preventing images of being indexed by search engines.
func NewNoImageIndexReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if strings.Contains(pageReport.Robots, "noimageindex") {
			return true
		}

		return false
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorNoImageIndex,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function to check
// if a page has missing img elements in Pictures.
func NewMissingImgTagInPictureReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		e := htmlquery.Find(htmlNode, "//picture")
		for _, n := range e {
			images := htmlquery.Find(n, "//img")
			if len(images) == 0 {
				return true
			}
		}

		return false
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorMissingImgElement,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function to check
// if a page has img elements without width or height attributes.
func NewImgWithoutSizeReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		e := htmlquery.Find(htmlNode, "//img")
		for _, n := range e {
			w := htmlquery.SelectAttr(n, "width")
			if w == "" {
				return true
			}

			h := htmlquery.SelectAttr(n, "height")
			if h == "" {
				return true
			}
		}

		return false
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorImgWithoutSize,
		Callback:  c,
	}
}
