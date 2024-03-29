package page

import (
	"net/http"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
)

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the page is not indexable by search engines.
func NewNoIndexableReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		return pageReport.Noindex
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorNoIndexable,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the page is blocked by the robots.txt file.
func NewBlockedByRobotstxtReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		return pageReport.BlockedByRobotstxt
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorBlocked,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the pageReport is non-indexable and it is included in the sitemap.
func NewNoIndexInSitemapReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		return pageReport.InSitemap && pageReport.Noindex
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorSitemapNoIndex,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the page is included in the sitemap and it is also blocked by the robots.txt file.
func NewSitemapAndBlockedReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		return pageReport.InSitemap && pageReport.BlockedByRobotstxt
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorSitemapBlocked,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the page is non canonical and it is included in the sitemap.
func NewNonCanonicalInSitemapReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.Canonical == "" || pageReport.Canonical == pageReport.URL {
			return false
		}

		return true
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorSitemapNonCanonical,
		Callback:  c,
	}
}
