package reporters

import (
	"net/url"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the hreflang values do not include an x-default option.
func NewHreflangXDefaultMissing() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if len(pageReport.Hreflangs) == 0 {
			return false
		}

		for _, hreflang := range pageReport.Hreflangs {
			if hreflang.Lang == "x-default" {
				return false
			}
		}

		return true
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorHreflangMissingXDefault,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the hreflang values don't include a self-referencing link.
func NewHreflangMissingSelfReference() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if len(pageReport.Hreflangs) == 0 {
			return false
		}

		for _, hl := range pageReport.Hreflangs {
			if hl.URL == pageReport.URL {
				return false
			}
		}

		return true
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorHreflangMissingSelfReference,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the self-referencing hreflang lang doesn't match the page's lang.
func NewHreflangMismatchingLang() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.Lang == "" {
			return false
		}

		for _, hl := range pageReport.Hreflangs {
			if hl.URL == pageReport.URL && hl.Lang != pageReport.Lang {
				return true
			}
		}

		return false
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorHreflangMismatchLang,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the hreflang URLs are relative.
func NewHreflangRelativeURL() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		hreflangs, err := htmlquery.QueryAll(htmlNode, "//head/link[@rel=\"alternate\"]")
		if err != nil || hreflangs == nil {
			return false
		}

		for _, hl := range hreflangs {
			parsedURL, err := url.Parse(htmlquery.SelectAttr(hl, "href"))
			if err != nil {
				return false
			}

			if !parsedURL.IsAbs() {
				return true
			}
		}

		return false
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorHreflangRelativeURL,
		Callback:  c,
	}
}
