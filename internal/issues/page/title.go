package page

import (
	"net/http"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// Returns a report_manager.PageIssueReporter with a callback function that checks if page has an empty little.
// The callback function returns true if the page is text/html, has a 20x status code
// and has an empty or missing title.
func NewEmptyTitleReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.StatusCode < 200 || pageReport.StatusCode >= 300 {
			return false
		}

		return pageReport.Title == ""
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorEmptyTitle,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that checks if the page has a short title.
// The callback returns true if the page is text/html and has a page title shorter than an specified
// amount of letters.
func NewShortTitleReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		return len(pageReport.Title) > 0 && len(pageReport.Title) < 20
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorShortTitle,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that checks if the page has a long title.
// The callback function returns true if the page is text/html and has a page title longer than an
// specified amount of letters.
func NewLongTitleReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		return len(pageReport.Title) > 60
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorLongTitle,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that checks if the page has more
// than one title tag in the header section.
// The callback returns true if the page is text/html and has more than one title in the header section.
func NewMultipleTitleTagsReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		tags, err := htmlquery.QueryAll(htmlNode, "//head/title")
		if err != nil {
			return false
		}

		return len(tags) > 1

	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorMultipleTitleTags,
		Callback:  c,
	}
}
