package page

import (
	"net/http"

	"github.com/antchfx/htmlquery"
	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/urlutils"

	"golang.org/x/net/html"
)

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the urls in the link rel pagination attributes do not exist as links in the body.
func NewPaginationReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.StatusCode >= 300 && pageReport.StatusCode < 400 {
			return false
		}

		rel, err := htmlquery.QueryAll(htmlNode, "//head/link[@rel=\"next\" or @rel=\"prev\" or @rel=\"previous\"]/@href")
		if err != nil {
			return false
		}

		linkMap := make(map[string]struct{}, len(pageReport.Links))
		for _, link := range pageReport.Links {
			linkMap[link.URL] = struct{}{}
		}

		for _, r := range rel {
			rh := htmlquery.SelectAttr(r, "href")
			arh, err := urlutils.AbsoluteURL(rh, htmlNode, pageReport.ParsedURL)
			if err != nil {
				continue
			}

			if _, exists := linkMap[arh.String()]; !exists {
				return true
			}
		}

		return false
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorPaginationLink,
		Callback:  c,
	}
}
