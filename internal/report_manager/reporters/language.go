package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
)

// Returns true if the page's html language is not valid.
func InvalidLang(pageReport *models.PageReport) bool {
	if pageReport.Crawled == false {
		return false
	}

	if pageReport.MediaType != "text/html" {
		return false
	}

	if pageReport.ValidLang == true {
		return false
	}

	return true
}

// Returns true if the page's html language is empty or missing.
func MissingLang(pageReport *models.PageReport) bool {
	if pageReport.Crawled == false {
		return false
	}

	if pageReport.MediaType != "text/html" {
		return false
	}

	return pageReport.Lang == ""
}
