package report

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
)

type PageReport struct {
	Id            int
	URL           string
	ParsedURL     *url.URL
	RedirectURL   string
	Refresh       string
	StatusCode    int
	ContentType   string
	MediaType     string
	Lang          string
	Title         string
	Description   string
	Robots        string
	Canonical     string
	H1            string
	H2            string
	Links         []Link
	ExternalLinks []Link
	Words         int
	Hreflangs     []Hreflang
	Body          []byte
	Size          int
	Images        []Image
	Scripts       []string
	Styles        []string
	sanitizer     *bluemonday.Policy
	ValidHeadings bool
}

type Link struct {
	URL       string
	ParsedURL *url.URL
	Rel       string
	Text      string
	External  bool
	NoFollow  bool
}

type Hreflang struct {
	URL  string
	Lang string
}

type Image struct {
	URL string
	Alt string
}

func NewPageReport(u *url.URL, status int, headers *http.Header, body []byte) *PageReport {
	pageReport := PageReport{
		URL:           u.String(),
		ParsedURL:     u,
		StatusCode:    status,
		ContentType:   headers.Get("Content-Type"),
		Body:          body,
		Size:          len(body),
		sanitizer:     bluemonday.StrictPolicy(),
		ValidHeadings: true,
	}

	mediaType, _, err := mime.ParseMediaType(pageReport.ContentType)
	if err != nil {
		log.Printf("NewPageReport: %v\n", err)
	}
	pageReport.MediaType = mediaType

	if pageReport.StatusCode >= http.StatusMultipleChoices && pageReport.StatusCode < http.StatusBadRequest {
		l, err := pageReport.absoluteURL(headers.Get("Location"))
		if err == nil {
			pageReport.RedirectURL = l.String()
		}
		return &pageReport
	}

	if mediaType == "text/html" {
		pageReport.parse()
	}

	return &pageReport
}

func (pageReport *PageReport) parse() {
	doc, err := htmlquery.Parse(bytes.NewReader(pageReport.Body))
	if err != nil {
		log.Printf("parse: %v\n", err)
		return
	}

	// ---
	// The lang attribute of the html element defines the document language
	// ex. <html lang="en">
	// ---
	lang := htmlquery.Find(doc, "//html/@lang")
	if len(lang) > 0 {
		pageReport.Lang = htmlquery.SelectAttr(lang[0], "lang")
	}

	// ---
	// The title element in the head section defines the page title
	// ex. <title>Test Page Title</title>
	// ---
	title := htmlquery.Find(doc, "//title")
	if len(title) > 0 {
		t := htmlquery.InnerText(title[0])
		pageReport.Title = strings.TrimSpace(t)
	}

	// ---
	// The description meta tag defines the page description
	// ex. <meta name="description" content="Test Page Description" />
	// ---
	description := htmlquery.Find(doc, "//meta[@name=\"description\"]/@content")
	if len(description) > 0 {
		d := htmlquery.SelectAttr(description[0], "content")
		pageReport.Description = strings.TrimSpace(d)
	}

	// ---
	// The refresh meta tag refreshes current page or redirects to a different one
	// ex. <meta http-equiv="refresh" content="0;URL='https://example.com/'" />
	// ---
	refresh := htmlquery.Find(doc, "//meta[@http-equiv=\"refresh\"]/@content")
	if len(refresh) > 0 {
		pageReport.Refresh = htmlquery.SelectAttr(refresh[0], "content")
		u := strings.Split(pageReport.Refresh, ";")
		if len(u) > 1 && strings.ToLower(u[1][:4]) == "url=" {
			l, err := pageReport.absoluteURL(strings.ReplaceAll(u[1][4:], "'", ""))
			if err == nil {
				pageReport.RedirectURL = l.String()
			}
		}
	}

	// ---
	// The robots meta provides information to crawlers
	// ex. <meta name="robots" content="noindex, nofollow" />
	// ---
	robots := htmlquery.Find(doc, "//meta[@name=\"robots\"]/@content")
	if len(robots) > 0 {
		pageReport.Robots = htmlquery.SelectAttr(robots[0], "content")
	}

	// ---
	// The a tags contain links to other pages we may want to crawl
	// ex. <a href="https://example.com/link1">link1</a>
	// ---
	list := htmlquery.Find(doc, "//a[@href]")
	for _, n := range list {
		l, err := pageReport.newLink(n)
		if err != nil {
			continue
		}

		if l.External {
			pageReport.ExternalLinks = append(pageReport.ExternalLinks, l)
		} else {
			pageReport.Links = append(pageReport.Links, l)
		}
	}

	// ---
	// H1 heading title
	// ex. <h1>H1 Title</h1>
	// ---
	h1 := htmlquery.Find(doc, "//h1")
	if len(h1) > 0 {
		pageReport.H1 = strings.TrimSpace(pageReport.sanitizer.Sanitize(htmlquery.InnerText(h1[0])))
	}

	// ---
	// H2 heading title
	// ex. <h2>H2 Title</h2>
	// ---
	h2 := htmlquery.Find(doc, "//h2")
	if len(h2) > 0 {
		pageReport.H2 = strings.TrimSpace(pageReport.sanitizer.Sanitize(htmlquery.InnerText(h2[0])))
	}

	// ---
	// Canonical link defines the main version for duplicate and similar pages
	// ex. <link rel="canonical" href="http://example.com/canonical/" />
	// ---
	canonical := htmlquery.Find(doc, "//link[@rel=\"canonical\"]/@href")
	if len(canonical) == 1 {
		cu, err := pageReport.absoluteURL(htmlquery.SelectAttr(canonical[0], "href"))
		if err == nil {
			pageReport.Canonical = cu.String()
		}
	}

	// ---
	// Extract hreflang urls so we can send them to the crawler
	// ex. <link rel="alternate" href="http://example.com" hreflang="am" />
	// ---
	hreflang := htmlquery.Find(doc, "//link[@rel=\"alternate\"]")
	for _, n := range hreflang {
		if htmlquery.ExistsAttr(n, "hreflang") {
			l, err := pageReport.absoluteURL(htmlquery.SelectAttr(n, "href"))
			if err != nil {
				continue
			}

			h := Hreflang{
				URL:  l.String(),
				Lang: htmlquery.SelectAttr(n, "hreflang"),
			}
			pageReport.Hreflangs = append(pageReport.Hreflangs, h)
		}
	}

	// ---
	// Extract images to check alt text and crawl src url
	// ex. <img src="logo.jpg">
	// ---
	images := htmlquery.Find(doc, "//img")
	for _, n := range images {
		s := htmlquery.SelectAttr(n, "src")
		url, err := pageReport.absoluteURL(s)
		if err != nil {
			continue
		}

		i := Image{
			URL: url.String(),
			Alt: htmlquery.SelectAttr(n, "alt"),
		}

		pageReport.Images = append(pageReport.Images, i)
	}

	// ---
	// Extract scripts to crawl the src url
	// ex. <script src="/js/app.js"></script>
	// ---
	scripts := htmlquery.Find(doc, "//script[@src]/@src")
	for _, n := range scripts {
		s := htmlquery.SelectAttr(n, "src")
		url, err := pageReport.absoluteURL(s)
		if err != nil {
			continue
		}

		pageReport.Scripts = append(pageReport.Scripts, url.String())
	}

	// ---
	// Extract stylesheet links to crawl the url
	// ex. <link rel="stylesheet" href="/css/style.css">
	// ---
	styles := htmlquery.Find(doc, "//link[@rel=\"stylesheet\"]/@href")
	for _, n := range styles {
		s := htmlquery.SelectAttr(n, "href")

		url, err := pageReport.absoluteURL(s)
		if err != nil {
			continue
		}

		pageReport.Styles = append(pageReport.Styles, url.String())
	}

	// ---
	// Count the words in the html body
	// ---
	body := htmlquery.Find(doc, "//body")
	if len(body) > 0 {
		pageReport.Words = countWords(body[0])
	}

	pageReport.ValidHeadings = headingOrderIsValid(body[0])
}

func (p *PageReport) newLink(n *html.Node) (Link, error) {
	href := htmlquery.SelectAttr(n, "href")
	u, err := p.absoluteURL(href)
	if err != nil {
		return Link{}, err
	}

	rel := strings.TrimSpace(htmlquery.SelectAttr(n, "rel"))

	l := Link{
		URL:       u.String(),
		ParsedURL: u,
		Rel:       rel,
		Text:      p.sanitizer.Sanitize(strings.TrimSpace(htmlquery.InnerText(n))),
		External:  u.Host != p.ParsedURL.Host,
		NoFollow:  strings.Contains(rel, "nofollow"),
	}

	return l, nil
}

func (p *PageReport) absoluteURL(s string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil {
		return &url.URL{}, err
	}

	if u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https" {
		return &url.URL{}, errors.New("Protocol not supported")
	}

	if u.Scheme != "" {
		if u.Path == "" {
			u.Path = "/"
		}
		return u, nil
	}

	if u.Scheme == "" {
		u.Scheme = p.ParsedURL.Scheme
	}

	if u.Host == "" {
		u.Host = p.ParsedURL.Host
	}

	u.Fragment = ""

	if u.Path != "" && !strings.HasPrefix(u.Path, "/") {
		basePath := p.ParsedURL.Path
		if !strings.HasSuffix(basePath, "/") {
			basePath = basePath + "/"
		}
		u.Path = basePath + u.Path
	}

	if u.Path == "" {
		basePath := p.ParsedURL.Path
		if basePath == "" {
			basePath = "/"
		}
		u.Path = basePath
	}

	return u, nil
}

func (p PageReport) SizeInKB() string {
	v := p.Size / (1 << 10)
	r := p.Size % (1 << 10)

	return fmt.Sprintf("%.2f", float64(v)+float64(r)/float64(1<<10))
}

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
