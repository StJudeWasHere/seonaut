package crawler

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
	"golang.org/x/net/html/charset"
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
	Body               []byte
	Headers            *http.Header
	Size               int
	Images             []Image
	Scripts            []string
	Styles             []string
	Iframes            []string
	Audios             []string
	Videos             []string
	sanitizer          *bluemonday.Policy
	ValidHeadings      bool
	BlockedByRobotstxt bool
	Crawled            bool
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

func NewPageReport(u *url.URL, status int, headers *http.Header, body []byte) *PageReport {
	pageReport := PageReport{
		URL:           u.String(),
		ParsedURL:     u,
		StatusCode:    status,
		ContentType:   headers.Get("Content-Type"),
		Body:          body,
		Headers:       headers,
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

	if pageReport.isHTML() {
		pageReport.parse()
	}

	return &pageReport
}

func (pageReport *PageReport) parse() {
	contentType := pageReport.ContentType
	utf8Body, err := charset.NewReader(bytes.NewReader(pageReport.Body), contentType)
	if err != nil {
		log.Printf("charset error %s: %v", contentType, err)
		return
	}

	doc, err := htmlquery.Parse(utf8Body)
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

	if pageReport.Lang == "" {
		languages := strings.Split(pageReport.Headers.Get("Content-Language"), ",")
		if len(languages) > 0 {
			pageReport.Lang = strings.TrimSpace(languages[0])
		}
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
		pageReport.Noindex = strings.Contains(pageReport.Robots, "noindex")
		pageReport.Nofollow = strings.Contains(pageReport.Robots, "nofollow")
	}

	// Check robots noindex and nofollow in the HTTP header
	robotsTag := pageReport.Headers.Get("X-Robots-Tag")
	if strings.Contains(robotsTag, "noindex") {
		pageReport.Noindex = true
	}

	if strings.Contains(robotsTag, "nofollow") {
		pageReport.Nofollow = true
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

	if pageReport.Canonical == "" {
		pageReport.parseCanonicalFromHeader()
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

	pageReport.parseHreflangsFromHeader()

	// ---
	// Extract images to check alt text and crawl src and srcset urls
	// ex. <img src="logo.jpg" srcset="/files/16870/new-york-skyline-wide.jpg 3724w">
	// ---
	images := htmlquery.Find(doc, "//img")
	for _, n := range images {
		s := htmlquery.SelectAttr(n, "src")
		if s == "" {
			continue
		}

		url, err := pageReport.absoluteURL(s)
		if err != nil {
			continue
		}

		alt := htmlquery.SelectAttr(n, "alt")
		i := Image{
			URL: url.String(),
			Alt: alt,
		}
		pageReport.Images = append(pageReport.Images, i)

		imageSet := parseSrcSet(htmlquery.SelectAttr(n, "srcset"))
		for _, s := range imageSet {
			url, err := pageReport.absoluteURL(s)
			if err != nil {
				continue
			}

			i := Image{
				URL: url.String(),
				Alt: alt,
			}
			pageReport.Images = append(pageReport.Images, i)
		}
	}

	// ---
	// Extract iframe URLs
	// ex.
	// <iframe height="500" width="500" src="http://example.com"></iframe>
	// ---
	iframes := htmlquery.Find(doc, "//iframe")
	for _, n := range iframes {
		s := htmlquery.SelectAttr(n, "src")
		if s == "" {
			continue
		}

		u, err := pageReport.absoluteURL(s)
		if err != nil {
			continue
		}

		pageReport.Iframes = append(pageReport.Iframes, u.String())
	}

	// ---
	// Extract image sources from picture elements.
	// The image's alt text is used in the sources.
	// ex.
	// <picture>
	//     <source srcset="/img/pic-wide.png" media="(min-width: 800px)">
	//     <source srcset="/img/pic-medium.png" media="(min-width: 600px)">
	//     <img src="/img/pic-narrow.png" alt="picture alt">
	// </picture>
	// ---
	pictures := htmlquery.Find(doc, "//picture")
	for _, n := range pictures {
		images := htmlquery.Find(n, "//img")
		if len(images) == 0 {
			continue
		}

		alt := htmlquery.SelectAttr(images[0], "alt")
		sources := htmlquery.Find(n, "//source")
		for _, s := range sources {
			imageSet := parseSrcSet(htmlquery.SelectAttr(s, "srcset"))
			for _, is := range imageSet {
				url, err := pageReport.absoluteURL(is)
				if err != nil {
					continue
				}

				i := Image{
					URL: url.String(),
					Alt: alt,
				}
				pageReport.Images = append(pageReport.Images, i)
			}
		}
	}

	// ---
	// Extract URLs from audio elements.
	// ex.
	// <audio src="audio_file.ogg" controls>
	// <source src="audio_file.wav" type="audio/wav">
	// </audio>
	// ---
	audios := htmlquery.Find(doc, "//audio")
	for _, n := range audios {

		src := htmlquery.SelectAttr(n, "src")
		if strings.TrimSpace(src) != "" {
			url, err := pageReport.absoluteURL(src)
			if err == nil {
				pageReport.Audios = append(pageReport.Audios, url.String())
			}
		}

		sources := htmlquery.Find(n, "//source")
		for _, s := range sources {
			src := htmlquery.SelectAttr(s, "src")
			url, err := pageReport.absoluteURL(src)
			if err != nil {
				continue
			}

			pageReport.Audios = append(pageReport.Audios, url.String())
		}
	}

	// ---
	// Extract URLs from video elements.
	// ex.
	// <video controls width="250">
	// <source src="video_file.webm" type="video/webm">
	// <source src="video_file.mp4" type="video/mp4">
	// </video>
	// ---
	videos := htmlquery.Find(doc, "//video")
	for _, n := range videos {

		src := htmlquery.SelectAttr(n, "src")
		if strings.TrimSpace(src) != "" {
			url, err := pageReport.absoluteURL(src)
			if err == nil {
				pageReport.Videos = append(pageReport.Videos, url.String())
			}
		}

		sources := htmlquery.Find(n, "//source")
		for _, s := range sources {
			src := htmlquery.SelectAttr(s, "src")
			url, err := pageReport.absoluteURL(src)
			if err != nil {
				continue
			}

			pageReport.Videos = append(pageReport.Videos, url.String())
		}
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

// Build a new link from a node element
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
		Sponsored: strings.Contains(rel, "sponsored"),
		UGC:       strings.Contains(rel, "ugc"),
	}

	return l, nil
}

// Return an absolute URL removing the URL fragment
func (p *PageReport) absoluteURL(s string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil {
		return &url.URL{}, err
	}

	a := p.ParsedURL.ResolveReference(u)
	a.Fragment = ""

	if a.Path == "" {
		a.Path = "/"
	}

	if a.Scheme != "http" && a.Scheme != "https" {
		return &url.URL{}, errors.New("Protocol not supported")
	}

	return a, nil
}

// Converts size KB and returns a string
func (p *PageReport) SizeInKB() string {
	v := p.Size / (1 << 10)
	r := p.Size % (1 << 10)

	return fmt.Sprintf("%.2f", float64(v)+float64(r)/float64(1<<10))
}

// Parse hreflang links from the HTTP header
func (p *PageReport) parseHreflangsFromHeader() {
	linkHeaderElements := strings.Split(p.Headers.Get("Link"), ",")
	for _, lh := range linkHeaderElements {
		attr := strings.Split(lh, ";")
		if len(attr) > 1 {
			url := strings.TrimSpace(attr[0])
			isAlternate := false
			lang := ""
			for _, a := range attr[1:] {
				a = strings.TrimSpace(a)
				if strings.Contains(a, `rel="alternate"`) {
					isAlternate = true
				}

				if strings.HasPrefix(a, "hreflang=") {
					lang = strings.Trim(a[9:], "\"")
				}
			}

			if isAlternate == true && lang != "" {
				h := Hreflang{
					URL:  url[1 : len(url)-1],
					Lang: lang,
				}

				p.Hreflangs = append(p.Hreflangs, h)
			}
		}
	}
}

// Parse hreflang links from the HTTP header
func (p *PageReport) parseCanonicalFromHeader() {
	linkHeaderElements := strings.Split(p.Headers.Get("Link"), ",")
	for _, lh := range linkHeaderElements {
		attr := strings.Split(lh, ";")
		if len(attr) == 2 && strings.Contains(attr[1], `rel="canonical"`) {
			url := strings.TrimSpace(attr[0])
			p.Canonical = url[1 : len(url)-1]

			return
		}
	}
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

// Parse srcset attribute and return the URLs
// ex. srcset="/img/image-wide.jpg 3724w,
//             /img/image-4by3.jpg 1961w,
//             /img/image-tall.jpg 1060w"
func parseSrcSet(srcset string) []string {
	var imageURLs []string

	if srcset == "" {
		return imageURLs
	}

	imageSet := strings.Split(srcset, ",")
	for _, s := range imageSet {
		i := strings.Split(s, " ")
		imageURLs = append(imageURLs, strings.TrimSpace(i[0]))
	}

	return imageURLs
}
