package reporters

import (
	"net/http"
	"strings"

	"golang.org/x/net/html"

	"github.com/antchfx/htmlquery"
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the page's html
// doesn't have any H1 tag.
func NewNoH1Reporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.StatusCode < 200 && pageReport.StatusCode >= 300 {
			return false
		}

		return pageReport.H1 == ""
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorNoH1,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the media type is text/html, the status code is between 200 and 299 and the heading tags
// in the page's html doesn't have the correct order.
func NewValidHeadingsOrderReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.StatusCode < 200 && pageReport.StatusCode >= 300 {
			return false
		}

		body, err := htmlquery.Query(htmlNode, "//body")
		if err != nil || body == nil {
			return false
		}

		headings := [6]string{"h1", "h2", "h3", "h4", "h5", "h6"}
		current := 0

		isValidHeading := func(el string) (int, bool) {
			el = strings.ToLower(el)
			for i, v := range headings {
				if v == el {
					return i, true
				}
			}

			return 0, false
		}

		var output func(n *html.Node) bool
		output = func(n *html.Node) bool {
			if n.Type == html.ElementNode {
				p, ok := isValidHeading(n.Data)
				if ok {
					if p > current+1 {
						return false
					}
					current = p
				}
			}

			for child := n.FirstChild; child != nil; child = child.NextSibling {
				if child.Type == html.ElementNode {
					if !output(child) {
						return false
					}
				}
			}

			return true
		}

		correct := output(body)

		return !correct
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorNotValidHeadings,
		Callback:  c,
	}
}
