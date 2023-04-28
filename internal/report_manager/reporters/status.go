package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
)

// Report if a page has an status code in the 30x range.
func Status30x(pageReport *models.PageReport) bool {
	if pageReport.Crawled == false {
		return false
	}

	return pageReport.StatusCode >= 300 && pageReport.StatusCode < 400
}

// Report if a page has an status code in the 40x range.
func Status40x(pageReport *models.PageReport) bool {
	if pageReport.Crawled == false {
		return false
	}

	return pageReport.StatusCode >= 400 && pageReport.StatusCode < 500
}

// Report if a page has an status code greater or equal to 500.
func Status50x(pageReport *models.PageReport) bool {
	if pageReport.Crawled == false {
		return false
	}

	return pageReport.StatusCode >= 500
}
