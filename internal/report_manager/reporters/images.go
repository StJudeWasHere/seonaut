package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
)

// Report if a page has images with no alt attribute.
func NoAltText(pageReport *models.PageReport) bool {
	if pageReport.Crawled == false {
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
