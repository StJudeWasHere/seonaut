package page

import (
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"

	"github.com/antchfx/htmlquery"
	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"
)

// Returns a report_manager.PageIssueReporter with a callback function that
// checks if a page has little content. The callback returns true if the page is text/html,
// has a 20x status code and less than a specified amount of words.
func NewLittleContentReporter() *models.PageIssueReporter {
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

		return pageReport.Words < 200
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorLittleContent,
		Callback:  c,
	}
}

func NewIncorrectMediaTypeReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if pageReport.MediaType == "" {
			return true
		}

		ext := filepath.Ext(pageReport.ParsedURL.Path)
		if ext == "" {
			ext = ".html"
		}

		// Allow both "application/javascript" and "text/javascript" as valid types.
		if ext == ".js" {
			return pageReport.MediaType != "application/javascript" && pageReport.MediaType != "text/javascript"
		}

		mimeType := mime.TypeByExtension(ext)
		mimeType = strings.Split(mimeType, ";")[0]

		if mimeType == "" {
			return false
		}

		return mimeType != pageReport.MediaType
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorIncorrectMediaType,
		Callback:  c,
	}
}

func NewDuplicatedIdReporter() *models.PageIssueReporter {
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

		e := htmlquery.Find(htmlNode, "//*[@id]")
		ids := make(map[string]bool)
		for _, n := range e {
			id := htmlquery.SelectAttr(n, "id")
			if id == "" {
				continue
			}

			if ids[id] {
				return true
			}

			ids[id] = true
		}

		return false
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorDuplicatedId,
		Callback:  c,
	}
}

// NewDOMSizeReporter returns a new reporter that returns true if the HTML document has
// more than a specified number of nodes. Otherwise it returns false.
func NewDOMSizeReporter(size int) *models.PageIssueReporter {
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

		if htmlNode.Type == html.ErrorNode {
			return false
		}

		nodes := htmlquery.Find(htmlNode, "//*")

		return len(nodes) > size
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorDOMSize,
		Callback:  c,
	}
}
