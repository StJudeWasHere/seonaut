package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
)

// Report if a page has an empty or missing description.
func EmptyDescription(pageReport *models.PageReport) bool {
	if pageReport.Crawled == false {
		return false
	}

	if pageReport.MediaType != "text/html" {
		return false
	}

	if pageReport.StatusCode < 200 || pageReport.StatusCode >= 300 {
		return false
	}

	return pageReport.Description == ""
}

// Report if a page has a short description.
// Returns true if the page is text/html and has a description of less than an specified
// amount of letters.
func ShortDescription(pageReport *models.PageReport) bool {
	if pageReport.Crawled == false {
		return false
	}

	if pageReport.MediaType != "text/html" {
		return false
	}

	if pageReport.StatusCode < 200 || pageReport.StatusCode >= 300 {
		return false
	}

	return len(pageReport.Description) > 0 && len(pageReport.Description) < 80
}

// Report if a page has long description.
func LongDescription(pageReport *models.PageReport) bool {
	if pageReport.Crawled == false {
		return false
	}

	if pageReport.MediaType != "text/html" {
		return false
	}

	if pageReport.StatusCode < 200 || pageReport.StatusCode >= 300 {
		return false
	}

	return len(pageReport.Description) > 160
}
