package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
)

// Returns a PageIssueReporter with a callback function that returns true if
// the page is not indexable by search engines.
func NewNoIndexableReporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		return pageReport.Noindex
	}

	return &PageIssueReporter{
		ErrorType: ErrorNoIndexable,
		Callback:  c,
	}
}

// Returns a PageIssueReporter with a callback function that returns true if
// the page is blocked by the robots.txt file.
func NewBlockedByRobotstxtReporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		return pageReport.BlockedByRobotstxt
	}

	return &PageIssueReporter{
		ErrorType: ErrorBlocked,
		Callback:  c,
	}
}

// Returns a PageIssueReporter with a callback function that returns true if
// the pageReport is non-indexable and it is included in the sitemap.
func NewNoIndexInSitemapReporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		return pageReport.Noindex && pageReport.BlockedByRobotstxt
	}

	return &PageIssueReporter{
		ErrorType: SitemapNoIndex,
		Callback:  c,
	}
}

// Returns a PageIssueReporter with a callback function that returns true if
// the page is included in the sitemap and it is also blocked by the robots.txt file.
func NewSitemapAndBlockedReporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		return pageReport.InSitemap && pageReport.BlockedByRobotstxt
	}

	return &PageIssueReporter{
		ErrorType: SitemapBlocked,
		Callback:  c,
	}
}

// Returns a PageIssueReporter with a callback function that returns true if
// the page is non canonical and it is included in the sitemap.
func NewNonCanonicalInSitemapReporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		if pageReport.Crawled == false {
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

	return &PageIssueReporter{
		ErrorType: SitemapNonCanonical,
		Callback:  c,
	}
}
