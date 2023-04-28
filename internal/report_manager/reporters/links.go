package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
)

// Returns true if the page has too many links
func TooManyLinks(pageReport *models.PageReport) bool {
	if pageReport.Crawled == false {
		return false
	}

	if pageReport.MediaType != "text/html" {
		return false
	}

	if pageReport.StatusCode < 200 || pageReport.StatusCode >= 300 {
		return false
	}

	return len(pageReport.Links) > 100
}

// Returns true if the page has internal nofollow links.
func InternalNoFollowLinks(pageReport *models.PageReport) bool {
	if pageReport.Crawled == false {
		return false
	}

	if pageReport.MediaType != "text/html" {
		return false
	}

	if pageReport.StatusCode < 200 || pageReport.StatusCode >= 300 {
		return false
	}

	for _, l := range pageReport.Links {
		if l.NoFollow == true {
			return true
		}
	}

	return false
}

// Returns true if the page has external links without the nofollow attribute.
func ExternalLinkWitoutNoFollow(pageReport *models.PageReport) bool {
	if pageReport.Crawled == false {
		return false
	}

	if pageReport.MediaType != "text/html" {
		return false
	}

	if pageReport.StatusCode < 200 || pageReport.StatusCode >= 300 {
		return false
	}

	for _, l := range pageReport.ExternalLinks {
		if l.NoFollow == false {
			return true
		}
	}

	return false
}

// Returns true if the page has http links.
func HTTPLinks(pageReport *models.PageReport) bool {
	if pageReport.Crawled == false {
		return false
	}

	if pageReport.MediaType != "text/html" {
		return false
	}

	if pageReport.StatusCode < 200 || pageReport.StatusCode >= 300 {
		return false
	}

	for _, l := range pageReport.Links {
		if l.ParsedURL.Scheme == "http" {
			return true
		}
	}

	return false
}
