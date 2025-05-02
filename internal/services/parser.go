package services

import (
	"bytes"
	"log"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/urlutils"

	"github.com/antchfx/htmlquery"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

type Parser struct {
	sanitizer *bluemonday.Policy
	doc       *html.Node
	ParsedURL *url.URL
	Headers   *http.Header
}

func newParser(url *url.URL, headers *http.Header, body []byte) (*Parser, error) {
	if len(body) == 0 {
		return &Parser{
			ParsedURL: url,
			Headers:   headers,
			doc:       &html.Node{},
		}, nil
	}

	utf8Body, err := charset.NewReader(bytes.NewReader(body), headers.Get("Content-Type"))
	if err != nil {
		log.Println(("utf8 reader error"))
		return nil, err
	}

	doc, err := html.ParseWithOptions(utf8Body, html.ParseOptionEnableScripting(false))
	if err != nil {
		log.Println(("htmlquery parse error"))
		return nil, err
	}

	return &Parser{
		sanitizer: bluemonday.StrictPolicy(),
		doc:       doc,
		ParsedURL: url,
		Headers:   headers,
	}, nil
}

// Returns the parsed body.
func (p *Parser) getHtmlNode() *html.Node {
	return p.doc
}

// Returns the document language.
// Returns the language defined in the HTML lang attribute if it is not empty
// otherwise it returns the language defined in the Content-Language headers.
func (p *Parser) lang() string {
	lang := p.htmlLang()
	if lang == "" {
		return p.headersLang()
	}

	return lang
}

// Returns the document robots settings.
// It will return the html meta robots content if it is not empty
// otherwise it returns the content of the X-Robots-Tag header.
func (p *Parser) robots() string {
	robots := p.htmlMetaRobots()
	if robots == "" {
		return p.headersRobots()
	}

	return robots
}

// Returns the document canonical settings.
// It returns the canonical defined in the canonical HTML tag if it is not empty
// otherwise it returns the canonical defined in the HTTP headers.
func (p *Parser) canonical() string {
	canonical := p.htmlCanonical()
	if canonical == "" {
		return p.headersCanonical()
	}

	return canonical
}

// Returns the document hreflangs.
// It first looks into the hreflang HTML tags, if empty it returns the hreflangs
// defined in the HTTP headers.
func (p *Parser) hreflangs() []models.Hreflang {
	hreflang := p.htmlHreflang()
	if len(hreflang) == 0 {
		return p.headersHreflangs()
	}

	return hreflang
}

// The lang attribute of the html element defines the document language
// ex. <html lang="en">
func (p *Parser) htmlLang() string {
	lang, err := htmlquery.Query(p.doc, "//html/@lang")
	if err != nil || lang == nil {
		return ""
	}

	return htmlquery.SelectAttr(lang, "lang")
}

// The title element in the head section defines the page title
// ex. <title>Test Page Title</title>
func (p *Parser) htmlTitle() string {
	title, err := htmlquery.Query(p.doc, "//head/title")
	if err != nil || title == nil {
		return ""
	}

	t := htmlquery.InnerText(title)

	return strings.TrimSpace(t)
}

// The description meta tag defines the page description
// ex. <meta name="description" content="Test Page Description" />
func (p *Parser) htmlMetaDescription() string {
	description, err := htmlquery.Query(p.doc, "//head/meta[@name=\"description\"]/@content")
	if err != nil || description == nil {
		return ""
	}

	d := htmlquery.SelectAttr(description, "content")

	return strings.TrimSpace(d)
}

// Returns the contents of the meta refresh tag.
// The refresh meta tag refreshes current page or redirects to a different one
// ex. <meta http-equiv="refresh" content="0;URL='https://example.com/'" />
func (p *Parser) htmlMetaRefresh() string {
	refresh, err := htmlquery.Query(p.doc, "//head//meta[@http-equiv=\"refresh\"]/@content")
	if err != nil || refresh == nil {
		return ""
	}

	return htmlquery.SelectAttr(refresh, "content")
}

// Returns the URL defined in the meta refresh tag.
// In case there isn't any url it returns an empty string.
func (p *Parser) htmlMetaRefreshURL() string {
	r := p.htmlMetaRefresh()
	u := strings.Split(r, ";")
	if len(u) > 1 && strings.ToLower(u[1][:4]) == "url=" {
		l := strings.ReplaceAll(u[1][4:], "'", "")
		redirect, err := urlutils.AbsoluteURL(l, p.doc, p.ParsedURL)
		if err != nil {
			return ""
		}

		return redirect.String()
	}

	return ""
}

// The robots meta provides information to crawlers
// ex. <meta name="robots" content="noindex, nofollow" />
func (p *Parser) htmlMetaRobots() string {
	robots, err := htmlquery.Query(p.doc, "//head/meta[@name=\"robots\"]/@content")
	if err != nil || robots == nil {
		return ""
	}

	return htmlquery.SelectAttr(robots, "content")
}

// H1 heading title
// ex. <h1>H1 Title</h1>
func (p *Parser) htmlH1() string {
	h1, err := htmlquery.Query(p.doc, "//h1")
	if err != nil || h1 == nil {
		return ""
	}

	return strings.TrimSpace(p.sanitizer.Sanitize(htmlquery.InnerText(h1)))
}

// H2 heading title
// ex. <h2>H2 Title</h2>
func (p *Parser) htmlH2() string {
	h2, err := htmlquery.Query(p.doc, "//h2")
	if err != nil || h2 == nil {
		return ""
	}

	return strings.TrimSpace(p.sanitizer.Sanitize(htmlquery.InnerText(h2)))
}

// Canonical link defines the main version for duplicate and similar pages
// ex. <link rel="canonical" href="http://example.com/canonical/" />
func (p *Parser) htmlCanonical() string {
	canonical, err := htmlquery.QueryAll(p.doc, "//head/link[@rel=\"canonical\"]/@href")
	if err != nil || canonical == nil || len(canonical) > 1 {
		return ""
	}

	cu, err := urlutils.AbsoluteURL(htmlquery.SelectAttr(canonical[0], "href"), p.doc, p.ParsedURL)
	if err != nil {
		return ""

	}

	return cu.String()
}

// The a tags contain links to other pages we may want to crawl
// ex. <a href="https://example.com/link1">link1</a>
func (p *Parser) htmlLinks() []models.Link {
	links := []models.Link{}
	htmlLinks, err := htmlquery.QueryAll(p.doc, "//a[@href]")
	if err != nil {
		return links
	}

	for _, v := range htmlLinks {
		l, err := p.newLink(v)
		if err != nil {
			continue
		}

		links = append(links, l)
	}

	return links
}

// Extract hreflang urls so we can send them to the crawler
// ex. <link rel="alternate" href="http://example.com" hreflang="am" />
func (p *Parser) htmlHreflang() []models.Hreflang {
	hreflangs := []models.Hreflang{}
	hl, err := htmlquery.QueryAll(p.doc, "//head/link[@rel=\"alternate\"]")
	if err != nil {
		return hreflangs
	}

	for _, n := range hl {
		if htmlquery.ExistsAttr(n, "hreflang") {
			l, err := urlutils.AbsoluteURL(htmlquery.SelectAttr(n, "href"), p.doc, p.ParsedURL)
			if err != nil {
				continue
			}

			h := models.Hreflang{
				URL:  l.String(),
				Lang: htmlquery.SelectAttr(n, "hreflang"),
			}
			hreflangs = append(hreflangs, h)
		}
	}

	return hreflangs
}

// Extract images to check alt text and crawl src and srcset urls
// ex. <img src="logo.jpg" srcset="/files/16870/new-york-skyline-wide.jpg 3724w">
func (p *Parser) htmlImages() []models.Image {
	images := []models.Image{}
	imgs := htmlquery.Find(p.doc, "//img")
	for _, n := range imgs {
		s := htmlquery.SelectAttr(n, "src")
		if s == "" {
			continue
		}

		url, err := urlutils.AbsoluteURL(s, p.doc, p.ParsedURL)
		if err != nil {
			continue
		}

		alt := htmlquery.SelectAttr(n, "alt")
		i := models.Image{
			URL: url.String(),
			Alt: alt,
		}
		images = append(images, i)

		imageSet := p.parseSrcSet(htmlquery.SelectAttr(n, "srcset"))
		for _, s := range imageSet {
			url, err := urlutils.AbsoluteURL(s, p.doc, p.ParsedURL)
			if err != nil {
				continue
			}

			i := models.Image{
				URL: url.String(),
				Alt: alt,
			}
			images = append(images, i)
		}
	}

	return images
}

// Extract iframe URLs
// ex. <iframe height="500" width="500" src="http://example.com"></iframe>
func (p *Parser) htmlIframes() []string {
	iframes := []string{}
	i := htmlquery.Find(p.doc, "//iframe")
	for _, n := range i {
		s := htmlquery.SelectAttr(n, "src")
		if s == "" {
			continue
		}

		u, err := urlutils.AbsoluteURL(s, p.doc, p.ParsedURL)
		if err != nil {
			continue
		}

		iframes = append(iframes, u.String())
	}

	return iframes
}

// Extract image sources from picture elements.
// The image's alt text is used in the sources.
// ex.
// <picture>
//
//	<source srcset="/img/pic-wide.png" media="(min-width: 800px)">
//	<source srcset="/img/pic-medium.png" media="(min-width: 600px)">
//	<img src="/img/pic-narrow.png" alt="picture alt">
//
// </picture>
func (p *Parser) htmlPictures() []models.Image {
	pictures := []models.Image{}
	e := htmlquery.Find(p.doc, "//picture")
	for _, n := range e {
		images := htmlquery.Find(n, "//img")
		if len(images) == 0 {
			continue
		}

		alt := htmlquery.SelectAttr(images[0], "alt")
		sources := htmlquery.Find(n, "//source")
		for _, s := range sources {
			imageSet := p.parseSrcSet(htmlquery.SelectAttr(s, "srcset"))
			for _, is := range imageSet {
				url, err := urlutils.AbsoluteURL(is, p.doc, p.ParsedURL)
				if err != nil {
					continue
				}

				i := models.Image{
					URL: url.String(),
					Alt: alt,
				}
				pictures = append(pictures, i)
			}
		}
	}

	return pictures
}

// Extract URLs from audio elements.
// ex.
// <audio src="audio_file.ogg" controls>
// <source src="audio_file.wav" type="audio/wav">
// </audio>
func (p *Parser) htmlAudios() []string {
	audios := []string{}
	a := htmlquery.Find(p.doc, "//audio")
	for _, n := range a {

		src := htmlquery.SelectAttr(n, "src")
		if strings.TrimSpace(src) != "" {
			url, err := urlutils.AbsoluteURL(src, p.doc, p.ParsedURL)
			if err == nil {
				audios = append(audios, url.String())
			}
		}

		sources := htmlquery.Find(n, "//source")
		for _, s := range sources {
			src := htmlquery.SelectAttr(s, "src")
			url, err := urlutils.AbsoluteURL(src, p.doc, p.ParsedURL)
			if err != nil {
				continue
			}

			audios = append(audios, url.String())
		}
	}

	return audios
}

// Extract URLs from video elements.
// ex.
// <video controls width="250">
// <source src="video_file.webm" type="video/webm">
// <source src="video_file.mp4" type="video/mp4">
// </video>
func (p *Parser) htmlVideos() []models.Video {
	videos := []models.Video{}
	v := htmlquery.Find(p.doc, "//video")
	for _, n := range v {
		poster := ""
		posterAttr := htmlquery.SelectAttr(n, "poster")
		if strings.TrimSpace(posterAttr) != "" {
			pURL, err := urlutils.AbsoluteURL(posterAttr, p.doc, p.ParsedURL)
			if err == nil {
				poster = pURL.String()
			}
		}

		src := htmlquery.SelectAttr(n, "src")
		if strings.TrimSpace(src) != "" {
			url, err := urlutils.AbsoluteURL(src, p.doc, p.ParsedURL)
			if err == nil {
				videos = append(videos, models.Video{URL: url.String(), Poster: poster})
			}
		}

		sources := htmlquery.Find(n, "//source")
		for _, s := range sources {
			src := htmlquery.SelectAttr(s, "src")
			url, err := urlutils.AbsoluteURL(src, p.doc, p.ParsedURL)
			if err != nil {
				continue
			}

			videos = append(videos, models.Video{URL: url.String(), Poster: poster})
		}
	}

	return videos
}

// Extract scripts to crawl the src url
// ex. <script src="/js/app.js"></script>
func (p *Parser) htmlScripts() []string {
	scripts := []string{}
	s := htmlquery.Find(p.doc, "//script[@src]/@src")
	for _, n := range s {
		s := htmlquery.SelectAttr(n, "src")
		url, err := urlutils.AbsoluteURL(s, p.doc, p.ParsedURL)
		if err != nil {
			continue
		}

		scripts = append(scripts, url.String())
	}

	return scripts
}

// Extract stylesheet links to crawl the url
// ex. <link rel="stylesheet" href="/css/style.css">
func (p *Parser) htmlStyles() []string {
	styles := []string{}
	s := htmlquery.Find(p.doc, "//link[@rel=\"stylesheet\"]/@href")
	for _, n := range s {
		s := htmlquery.SelectAttr(n, "href")

		url, err := urlutils.AbsoluteURL(s, p.doc, p.ParsedURL)
		if err != nil {
			continue
		}

		styles = append(styles, url.String())
	}

	return styles
}

// Return the html document
// ex. <body>
func (p *Parser) htmlBodyNode() *html.Node {
	body, err := htmlquery.Query(p.doc, "//body")
	if err != nil {
		return nil
	}

	return body
}

// Parse hreflang links from the HTTP header
func (p *Parser) headersCanonical() string {
	linkHeaderElements := strings.Split(p.Headers.Get("Link"), ",")
	for _, lh := range linkHeaderElements {
		attr := strings.Split(lh, ";")
		if len(attr) == 2 && strings.Contains(attr[1], `rel="canonical"`) {
			canonicalString := strings.TrimSpace(attr[0])
			cu, err := urlutils.AbsoluteURL(canonicalString[1:len(canonicalString)-1], p.doc, p.ParsedURL)
			if err != nil {
				return ""
			}

			return cu.String()
		}
	}

	return ""
}

// Parse hreflang links from the HTTP header
func (p *Parser) headersHreflangs() []models.Hreflang {
	hreflangs := []models.Hreflang{}
	linkHeaderElements := strings.Split(p.Headers.Get("Link"), ",")
	for _, lh := range linkHeaderElements {
		attr := strings.Split(lh, ";")
		if len(attr) < 1 {
			continue
		}

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

		if isAlternate && lang != "" {
			h := models.Hreflang{
				URL:  url[1 : len(url)-1],
				Lang: lang,
			}

			hreflangs = append(hreflangs, h)
		}
	}

	return hreflangs
}

// Returns the language specified in the Content-Language headers.
func (p *Parser) headersLang() string {
	languages := strings.Split(p.Headers.Get("Content-Language"), ",")
	if len(languages) > 0 {
		return strings.TrimSpace(languages[0])
	}

	return ""
}

// Returns the contents of the X-Robots-Tag header.
func (p *Parser) headersRobots() string {
	return p.Headers.Get("X-Robots-Tag")
}

// Return the contents of the HTTP Location header.
func (p *Parser) headersLocation() string {
	l, err := urlutils.AbsoluteURL(p.Headers.Get("Location"), p.doc, p.ParsedURL)
	if err != nil {
		return ""
	}

	return l.String()
}

// Parse srcset attribute and return the URLs
// ex. srcset="/img/image-wide.jpg 3724w,
//
//	/img/image-4by3.jpg 1961w,
//	/img/image-tall.jpg 1060w"
func (p *Parser) parseSrcSet(srcset string) []string {
	var imageURLs []string

	srcset = strings.Trim(srcset, " ,")
	if srcset == "" {
		return imageURLs
	}

	// URLs in srcset strings can contain an optional descriptor.
	// Also take into account URLs with commas in them.
	parsingURL := true
	var currentURL strings.Builder
	for _, char := range srcset {
		if parsingURL {
			if unicode.IsSpace(char) {
				if currentURL.Len() > 0 {
					parsingURL = false
				}
			} else if currentURL.Len() > 0 || char != ',' {
				currentURL.WriteRune(char)
			}
		} else {
			if char == ',' {
				parsingURL = true
				imageURLs = append(imageURLs, strings.TrimSpace(currentURL.String()))
				currentURL.Reset()
			}
		}
	}

	if currentURL.Len() > 0 {
		imageURLs = append(imageURLs, strings.TrimSpace(currentURL.String()))
	}

	return imageURLs
}

// Build a new link from a node element
func (p *Parser) newLink(n *html.Node) (models.Link, error) {
	href := htmlquery.SelectAttr(n, "href")
	u, err := urlutils.AbsoluteURL(href, p.doc, p.ParsedURL)
	if err != nil {
		return models.Link{}, err
	}

	rel := strings.TrimSpace(htmlquery.SelectAttr(n, "rel"))

	l := models.Link{
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
