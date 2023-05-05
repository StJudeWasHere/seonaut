package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the page is not indexable by search engines.
func NewNoIndexableReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		return pageReport.Noindex
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorNoIndexable,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the page is blocked by the robots.txt file.
func NewBlockedByRobotstxtReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		return pageReport.BlockedByRobotstxt
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorBlocked,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the pageReport is non-indexable and it is included in the sitemap.
func NewNoIndexInSitemapReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		return pageReport.InSitemap && pageReport.Noindex
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorSitemapNoIndex,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the page is included in the sitemap and it is also blocked by the robots.txt file.
func NewSitemapAndBlockedReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
		return pageReport.InSitemap && pageReport.BlockedByRobotstxt
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorSitemapBlocked,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the page is non canonical and it is included in the sitemap.
func NewNonCanonicalInSitemapReporter() *report_manager.PageIssueReporter {
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

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorSitemapNonCanonical,
		Callback:  c,
	}
}
