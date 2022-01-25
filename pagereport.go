package main

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
	"golang.org/x/net/html"
)

type PageReport struct {
	Id            int
	URL           string
	parsedURL     *url.URL
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
}

type Link struct {
	URL      string
	Rel      string
	Text     string
	External bool
}

type Hreflang struct {
	URL  string
	Lang string
}

type Image struct {
	URL string
	Alt string
}

func NewPageReport(url *url.URL, status int, headers *http.Header, body []byte) *PageReport {
	pageReport := PageReport{
		URL:         url.String(),
		parsedURL:   url,
		StatusCode:  status,
		ContentType: headers.Get("Content-Type"),
		Body:        body,
		Size:        len(body),
	}

	mediaType, _, err := mime.ParseMediaType(pageReport.ContentType)
	if err != nil {
		fmt.Println(err)
	}
	pageReport.MediaType = mediaType

	if pageReport.StatusCode >= http.StatusMultipleChoices && pageReport.StatusCode < http.StatusBadRequest {
		pageReport.RedirectURL = headers.Get("Location")
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
		fmt.Println(err)
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
			pageReport.RedirectURL = strings.ReplaceAll(u[1][4:], "'", "")
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
		pageReport.H1 = strings.TrimSpace(htmlquery.InnerText(h1[0]))
	}

	// ---
	// H2 heading title
	// ex. <h2>H2 Title</h2>
	// ---
	h2 := htmlquery.Find(doc, "//h2")
	if len(h2) > 0 {
		pageReport.H2 = strings.TrimSpace(htmlquery.InnerText(h2[0]))
	}

	// ---
	// Canonical link defines the main version for duplicate and similar pages
	// ex. <link rel="canonical" href="http://example.com/canonical/" />
	// ---
	canonical := htmlquery.Find(doc, "//link[@rel=\"canonical\"]/@href")
	if len(canonical) == 1 {
		pageReport.Canonical = htmlquery.SelectAttr(canonical[0], "href")
	}

	// ---
	// Extract hreflang urls so we can send them to the crawler
	// ex. <link rel="alternate" href="http://example.com" hreflang="am" />
	// ---
	hreflang := htmlquery.Find(doc, "//link[@rel=\"alternate\"]")
	for _, n := range hreflang {
		if htmlquery.ExistsAttr(n, "hreflang") {
			h := Hreflang{
				URL:  htmlquery.SelectAttr(n, "href"),
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
		url, err := url.Parse(htmlquery.SelectAttr(n, "src"))
		if err != nil {
			log.Println(err)
			continue
		}

		url = pageReport.absoluteURL(url)
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
		url, err := url.Parse(htmlquery.SelectAttr(n, "src"))
		if err != nil {
			log.Println(err)
			continue
		}

		url = pageReport.absoluteURL(url)
		pageReport.Scripts = append(pageReport.Scripts, url.String())
	}

	// ---
	// Extract stylesheet links to crawl the url
	// ex. <link rel="stylesheet" href="/css/style.css">
	// ---
	styles := htmlquery.Find(doc, "//link[@rel=\"stylesheet\"]/@href")
	for _, n := range styles {
		url, err := url.Parse(htmlquery.SelectAttr(n, "href"))
		if err != nil {
			log.Println(err)
			continue
		}

		url = pageReport.absoluteURL(url)
		pageReport.Styles = append(pageReport.Styles, url.String())
	}

	// ---
	// Count the words in the html body
	// ---
	body := htmlquery.Find(doc, "//body")
	if len(body) > 0 {
		pageReport.Words = countWords(body[0])
	}
}

func (p *PageReport) newLink(n *html.Node) (Link, error) {
	href := htmlquery.SelectAttr(n, "href")
	u, err := url.Parse(href)
	if err != nil {
		return Link{}, err
	}

	if u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https" {
		return Link{}, errors.New("Protocol not supported")
	}

	u = p.absoluteURL(u)

	l := Link{
		URL:      u.String(),
		Rel:      htmlquery.SelectAttr(n, "rel"),
		Text:     htmlquery.InnerText(n),
		External: u.Host != p.parsedURL.Host,
	}

	return l, nil
}

func (p *PageReport) absoluteURL(u *url.URL) *url.URL {
	if u.Scheme == "" {
		u.Scheme = p.parsedURL.Scheme
	}

	if u.Host == "" {
		u.Host = p.parsedURL.Host
	}

	if !strings.HasPrefix(u.Path, "/") {
		basePath := p.parsedURL.Path
		if !strings.HasSuffix(basePath, "/") {
			basePath = basePath + "/"
		}
		u.Path = basePath + u.Path
	}

	return u
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
		fmt.Println(err)
	}
	t := re.ReplaceAllString(buf.String(), " ")

	return len(strings.Fields(t))
}
