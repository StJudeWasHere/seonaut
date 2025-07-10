package page

import (
	"net/http"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
)

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// contains too many links.
func NewTooManyLinksReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
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

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorTooManyLinks,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// contains internal links with the nofollow attribute.
func NewInternalNoFollowLinksReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.StatusCode < 200 || pageReport.StatusCode >= 300 {
			return false
		}

		if pageReport.Nofollow && len(pageReport.InternalLinks) > 0 {
			return true
		}

		for _, l := range pageReport.InternalLinks {
			if l.Link.NoFollow {
				return true
			}
		}

		return false
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorInternalNoFollow,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// contains external links without the nofollow attribute.
func NewExternalLinkWitoutNoFollowReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.StatusCode < 200 || pageReport.StatusCode >= 300 {
			return false
		}

		if pageReport.Nofollow {
			return false
		}

		for _, l := range pageReport.ExternalLinks {
			if !l.NoFollow {
				return true
			}
		}

		return false
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorExternalWithoutNoFollow,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// contains internal links with the http scheme instead of https.
func NewHTTPLinksReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
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

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorHTTPLinks,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// contains no internal or external links.
func NewDeadendReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.StatusCode < 200 || pageReport.StatusCode >= 300 {
			return false
		}

		return len(pageReport.Links)+len(pageReport.ExternalLinks) == 0
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorDeadend,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the media type is text/html, and contains external links with status code is between 300 and 399.
func NewExternalLinkRedirectReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		for _, l := range pageReport.ExternalLinks {
			if l.StatusCode >= 300 && l.StatusCode <= 399 {
				return true
			}
		}

		return false
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorExternalLinkRedirect,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the media type is text/html, and contains external links with status code is greater than 399.
func NewExternalLinkBrokenReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		for _, l := range pageReport.ExternalLinks {
			if l.StatusCode < 0 || l.StatusCode > 399 {
				return true
			}
		}

		return false
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorExternalLinkBroken,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the pageReport contains external links to localhost or 127.0.0.1.
func NewLocalhostLinksReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.ParsedURL.Host == "localhost" || pageReport.ParsedURL.Host == "127.0.0.1" {
			return false
		}

		for _, l := range pageReport.ExternalLinks {
			if l.ParsedURL.Host == "localhost" || l.ParsedURL.Host == "127.0.0.1" {
				return true
			}
		}

		return false
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorLocalhostLinks,
		Callback:  c,
	}
}
