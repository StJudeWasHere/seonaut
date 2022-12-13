package pagereport

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

const (
	// MaxBodySize is the limit of the retrieved response body in bytes.
	// The default value for MaxBodySize is 10MB (10 * 1024 * 1024 bytes).
	maxBodySize = 10 * 1024 * 1024
)

type PageReport struct {
	Id                 int
	URL                string
	ParsedURL          *url.URL
	RedirectURL        string
	Refresh            string
	StatusCode         int
	ContentType        string
	MediaType          string
	Lang               string
	Title              string
	Description        string
	Robots             string
	Noindex            bool
	Nofollow           bool
	Canonical          string
	H1                 string
	H2                 string
	Links              []Link
	ExternalLinks      []Link
	Words              int
	Hreflangs          []Hreflang
	Size               int
	Images             []Image
	Scripts            []string
	Styles             []string
	Iframes            []string
	Audios             []string
	Videos             []string
	ValidHeadings      bool
	BlockedByRobotstxt bool
	Crawled            bool
	InSitemap          bool
}

type Link struct {
	URL       string
	ParsedURL *url.URL
	Rel       string
	Text      string
	External  bool
	NoFollow  bool
	Sponsored bool
	UGC       bool
}

type Hreflang struct {
	URL  string
	Lang string
}

type Image struct {
	URL string
	Alt string
}

// Create a new PageReport from an http.Response
func NewPageReportFromHTTPResponse(r *http.Response) (*PageReport, error) {
	defer r.Body.Close()

	var bodyReader io.Reader = r.Body
	bodyReader = io.LimitReader(bodyReader, int64(maxBodySize))

	b, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return &PageReport{}, err
	}

	return NewPageReport(r.Request.URL, r.StatusCode, &r.Header, b)
}

// Return a new PageReport.
func NewPageReport(u *url.URL, status int, headers *http.Header, body []byte) (*PageReport, error) {
	parser, err := newParser(u, headers, body)
	if err != nil {
		return nil, err
	}

	pageReport := PageReport{
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

		return &pageReport, nil
	}

	if pageReport.isHTML() {
		pageReport.Lang = parser.lang()
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

		pageReport.Words = countWords(parser.htmlBodyNode())
		pageReport.ValidHeadings = headingOrderIsValid(parser.htmlBodyNode())
	}

	return &pageReport, nil
}

// Converts size KB and returns a string
func (p *PageReport) SizeInKB() string {
	v := p.Size / (1 << 10)
	r := p.Size % (1 << 10)

	return fmt.Sprintf("%.2f", float64(v)+float64(r)/float64(1<<10))
}

// Returns true if ContentType is a valid HTML type
func (p *PageReport) isHTML() bool {
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

// Check if the H headings order is valid
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
				if output(child) == false {
					return false
				}
			}
		}

		return true
	}

	correct := output(n)

	return correct
}
