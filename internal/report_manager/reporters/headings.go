package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
)

// Report if a page has an empty or missing H1 tag.
func NoH1(pageReport *models.PageReport) bool {
	if pageReport.Crawled == false {
		return false
	}

	if pageReport.MediaType != "text/html" {
		return false
	}

	if pageReport.StatusCode < 200 && pageReport.StatusCode >= 300 {
		return false
	}

	return pageReport.H1 == ""
}

// Returns true if HTML headings order is not valid
func ValidHeadingsOrder(pageReport *models.PageReport) bool {
	if pageReport.Crawled == false {
		return false
	}

	if pageReport.MediaType != "text/html" {
		return false
	}

	if pageReport.StatusCode < 200 && pageReport.StatusCode >= 300 {
		return false
	}

	if pageReport.ValidHeadings == true {
		return false
	}

	return true
}
