package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
)

// Report if page has an empty little.
// It returns true if the page is text/html, has a 20x status code
// and has an empty or missing title.
func EmptyTitle(pageReport *models.PageReport) bool {
	if pageReport.Crawled == false {
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

// Report if the page has a short title.
// It returns true if the page is text/html and has a page title shorter than an specified amount of letters.
func ShortTitle(pageReport *models.PageReport) bool {
	if pageReport.Crawled == false {
		return false
	}

	if pageReport.MediaType != "text/html" {
		return false
	}

	return len(pageReport.Title) > 0 && len(pageReport.Title) < 20
}

// Report if the page has a long title.
// It returns true if the page is text/html and has a page title longer than an specified amount of letters.
func LongTitle(pageReport *models.PageReport) bool {
	if pageReport.Crawled == false {
		return false
	}

	if pageReport.MediaType != "text/html" {
		return false
	}

	return len(pageReport.Title) > 60
}
