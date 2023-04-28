package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
)

// Report if page has little content.
// It returns true if the page is text/html, has a 20x status code
// and less than a specified amount of words.
func LittleContent(pageReport *models.PageReport) bool {
	if pageReport.Crawled == false {
		return false
	}

	if pageReport.MediaType != "text/html" {
		return false
	}

	if pageReport.StatusCode < 200 || pageReport.StatusCode >= 300 {
		return false
	}

	return pageReport.Words < 200
}
