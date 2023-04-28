package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
)

// Returns true if the page is not indexable by search engines.
func NoIndexable(pageReport *models.PageReport) bool {
	return pageReport.Noindex
}

// Returns true if the page is blocked by tge robots.txt file.
func BlockedByRobotstxt(pageReport *models.PageReport) bool {
	return pageReport.BlockedByRobotstxt
}

// Returns true if the pageReport is non-indexable and is included in the sitemap.
func NoIndexInSitemap(pageReport *models.PageReport) bool {
	return pageReport.Noindex && pageReport.BlockedByRobotstxt
}

// Returns true if the page is included in the sitemap and it is also blocked by the robots.txt file
func SitemapAndBlocked(pageReport *models.PageReport) bool {
	return pageReport.InSitemap && pageReport.BlockedByRobotstxt
}

// Returns true if the page is non canonical and it is included in the sitemap.
func NonCanonicalInSitemap(pageReport *models.PageReport) bool {
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
