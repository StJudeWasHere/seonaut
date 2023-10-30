package reporters

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// Returns a report_manager.PageIssueReporter with a callback function that
// reports if the page's HSTS header is missing. The callback returns true if the Strict-Transport-Security,
// header does not exist or is not valid.
func NewMissingHSTSHeaderReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		hstsHeader := header.Get("Strict-Transport-Security")
		if hstsHeader == "" {
			return true
		}

		directives := strings.Split(hstsHeader, ";")
		for _, directive := range directives {
			if strings.HasPrefix(directive, "max-age=") {
				maxAge := strings.TrimPrefix(directive, "max-age=")
				_, err := strconv.Atoi(maxAge)
				if err != nil {
					return true
				}
			}
		}

		return false
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorMissingHSTSHeader,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that
// reports if the page's CSP (Content Security Policy) is missing by looking both in the Headers and meta tags.
// The callback returns true if the CSP does not exist.
func NewMissingCSPReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if pageReport.MediaType != "text/html" {
			return false
		}

		cspTag, err := htmlquery.QueryAll(htmlNode, "//head/meta[@http-equiv=\"Content-Security-Policy\"]")
		if err != nil {
			return false
		}

		CSPHeader := header.Get("Content-Security-Policy")

		return cspTag == nil && CSPHeader == ""
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorMissingCSP,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that
// reports if the page's X-Content-Type-Options header is missing.
// The callback returns true if the header does not exist.
func NewMissingContentTypeOptionsReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		contentTypeOptions := header.Get("X-Content-Type-Options")

		return contentTypeOptions != "nosniff"
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorContentTypeOptions,
		Callback:  c,
	}
}
