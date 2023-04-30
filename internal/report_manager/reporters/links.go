package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
)

// Returns a PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// contains too many links.
func NewTooManyLinksReporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
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

	return &PageIssueReporter{
		ErrorType: ErrorTooManyLinks,
		Callback:  c,
	}
}

// Returns a PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// contains internal links with the nofollow attribute.
func NewInternalNoFollowLinksReporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
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

	return &PageIssueReporter{
		ErrorType: ErrorInternalNoFollow,
		Callback:  c,
	}
}

// Returns a PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// contains external links without the nofollow attribute.
func NewExternalLinkWitoutNoFollowReporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
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

	return &PageIssueReporter{
		ErrorType: ErrorExternalWithoutNoFollow,
		Callback:  c,
	}
}

// Returns a PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// contains internal links with the http scheme instead of https.
func NewHTTPLinksReporter() *PageIssueReporter {
	c := func(pageReport *models.PageReport) bool {
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

	return &PageIssueReporter{
		ErrorType: ErrorHTTPLinks,
		Callback:  c,
	}
}
