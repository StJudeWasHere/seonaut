package html_parser

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
	"golang.org/x/text/language"
)

const (
	// MaxBodySize is the limit of the retrieved response body in bytes.
	// The default value for MaxBodySize is 10MB (10 * 1024 * 1024 bytes).
	maxBodySize = 10 * 1024 * 1024
)

// Create a new PageReport from an http.Response.
func NewFromHTTPResponse(r *http.Response) (*models.PageReport, *html.Node, error) {
	defer r.Body.Close()

	var bodyReader io.Reader = r.Body
	bodyReader = io.LimitReader(bodyReader, int64(maxBodySize))

	b, err := io.ReadAll(bodyReader)
	if err != nil {
		return &models.PageReport{}, nil, err
	}

	return New(r.Request.URL, r.StatusCode, &r.Header, b)
}

// Return a new PageReport.
func New(u *url.URL, status int, headers *http.Header, body []byte) (*models.PageReport, *html.Node, error) {
	parser, err := newParser(u, headers, body)
	if err != nil {
		return nil, nil, err
	}

	pageReport := models.PageReport{
		URL:           u.String(),
		ParsedURL:     u,
		StatusCode:    status,
		ContentType:   headers.Get("Content-Type"),
		Size:          len(body),
		ValidHeadings: true,
	}

	pageReport.MediaType, _, err = mime.ParseMediaType(pageReport.ContentType)
	if err != nil {
		log.Printf("NewPageReport URL: %s\n Error: %v", u.String(), err)
	}

	if pageReport.StatusCode >= http.StatusMultipleChoices && pageReport.StatusCode < http.StatusBadRequest {
		pageReport.RedirectURL = parser.headersLocation()

		return &pageReport, parser.getHtmlNode(), nil
	}

	if isHTML(&pageReport) {
		pageReport.Lang = parser.lang()
		pageReport.ValidLang = langIsValid(pageReport.Lang)
		pageReport.Title = parser.htmlTitle()
		pageReport.Description = parser.htmlMetaDescription()
		pageReport.Refresh = parser.htmlMetaRefresh()
		pageReport.RedirectURL = parser.htmlMetaRefreshURL()
		pageReport.Robots = parser.robots()
		pageReport.Noindex = strings.Contains(pageReport.Robots, "noindex")
		pageReport.Nofollow = strings.Contains(pageReport.Robots, "nofollow")
		pageReport.H1 = parser.htmlH1()
		pageReport.H2 = parser.htmlH2()
		pageReport.Canonical = parser.canonical()
		pageReport.Hreflangs = parser.hreflangs()
		pageReport.Images = parser.htmlImages()
		pageReport.Iframes = parser.htmlIframes()
		pageReport.Audios = parser.htmlAudios()
		pageReport.Videos = parser.htmlVideos()
		pageReport.Scripts = parser.htmlScripts()
		pageReport.Styles = parser.htmlStyles()

		pictures := parser.htmlPictures()
		pageReport.Images = append(pageReport.Images, pictures...)

		links := parser.htmlLinks()
		for _, l := range links {
			if l.External {
				pageReport.ExternalLinks = append(pageReport.ExternalLinks, l)
			} else {
				pageReport.Links = append(pageReport.Links, l)
			}
		}

		bnode := parser.htmlBodyNode()
		if bnode != nil {
			pageReport.Words = countWords(bnode)
			pageReport.ValidHeadings = headingOrderIsValid(bnode)
		}
	}

	return &pageReport, parser.getHtmlNode(), nil
}

// Returns true if ContentType is a valid HTML type
func isHTML(p *models.PageReport) bool {
	validTypes := []string{"text/html", "application/xhtml+xml", "application/vnd.wap.xhtml+xml"}
	for _, t := range validTypes {
		if strings.Contains(p.ContentType, t) {
			return true
		}
	}

	return false
}

// Count number of words in an HTML node
func countWords(n *html.Node) int {
	var output func(*bytes.Buffer, *html.Node)
	output = func(buf *bytes.Buffer, n *html.Node) {
		switch n.Type {
		case html.TextNode:
			if n.Parent.Type == html.ElementNode && n.Parent.Data != "script" {
				buf.WriteString(fmt.Sprintf("%s ", n.Data))
			}
			return
		case html.CommentNode:
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if child.Parent.Type == html.ElementNode && child.Parent.Data != "a" {
				output(buf, child)
			}
		}
	}

	var buf bytes.Buffer
	output(&buf, n)

	re, err := regexp.Compile(`[\p{P}\p{S}]+`)
	if err != nil {
		log.Printf("countWords: %v\n", err)
	}
	t := re.ReplaceAllString(buf.String(), " ")

	return len(strings.Fields(t))
}

// Check if the H headings order is valid.
func headingOrderIsValid(n *html.Node) bool {
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

	correct := output(n)

	return correct
}

// Check if a language code provided by the Content-Language header or HTML lang attribute is valid.
func langIsValid(s string) bool {
	langs := strings.Split(s, ",")
	for _, l := range langs {
		_, err := language.Parse(l)
		if err != nil {
			return false
		}
	}

	return true
}
