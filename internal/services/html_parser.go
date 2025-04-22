package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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
)

const (
	// MaxBodySize is the limit of the retrieved response body in bytes.
	// The default value for MaxBodySize is 10MB (10 * 1024 * 1024 bytes).
	maxBodySize = 10 * 1024 * 1024
)

// Create a new PageReport from an http.Response.
func NewFromHTTPResponse(r *http.Response) (*models.PageReport, *html.Node, error) {
	defer r.Body.Close()

	var bodyCopy bytes.Buffer
	_, err := io.Copy(&bodyCopy, r.Body)
	if err != nil {
		return &models.PageReport{}, &html.Node{}, err
	}

	r.Body = io.NopCloser(bytes.NewReader(bodyCopy.Bytes()))

	var bodyReader io.Reader = bytes.NewReader(bodyCopy.Bytes())
	bodyReader = io.LimitReader(bodyReader, int64(maxBodySize))

	b, err := io.ReadAll(bodyReader)
	if err != nil {
		return &models.PageReport{}, &html.Node{}, err
	}

	return NewHTMLParser(r.Request.URL, r.StatusCode, &r.Header, b, r.ContentLength)
}

// Return a new PageReport.
func NewHTMLParser(u *url.URL, status int, headers *http.Header, body []byte, contentLength int64) (*models.PageReport, *html.Node, error) {
	parser, err := newParser(u, headers, body)
	if err != nil {
		log.Println("newParser error!")
		return &models.PageReport{}, &html.Node{}, err
	}

	size := int64(len(body))
	if size == 0 && contentLength > size {
		size = contentLength
	}

	pageReport := models.PageReport{
		URL:         u.String(),
		ParsedURL:   u,
		StatusCode:  status,
		ContentType: headers.Get("Content-Type"),
		Size:        size,
	}

	pageReport.MediaType, _, _ = mime.ParseMediaType(pageReport.ContentType)

	if pageReport.StatusCode >= http.StatusMultipleChoices && pageReport.StatusCode < http.StatusBadRequest {
		pageReport.RedirectURL = parser.headersLocation()

		return &pageReport, parser.getHtmlNode(), nil
	}

	if isHTML(&pageReport) && size > 0 {
		pageReport.Lang = parser.lang()
		pageReport.Title = parser.htmlTitle()
		pageReport.Description = parser.htmlMetaDescription()
		pageReport.Refresh = parser.htmlMetaRefresh()
		pageReport.RedirectURL = parser.htmlMetaRefreshURL()
		pageReport.Robots = parser.robots()
		pageReport.Noindex = containsAny(pageReport.Robots, "noindex", "none")
		pageReport.Nofollow = containsAny(pageReport.Robots, "nofollow", "none")
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
		}

		pageReport.BodyHash, err = hashString(body)
		if err != nil {
			log.Printf("body hashString URL: %s\nError %v", u.String(), err)
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

// Hash a string using sha256 and returns is hex representation as a string.
func hashString(input []byte) (string, error) {
	hasher := sha256.New()
	_, err := hasher.Write(input)
	if err != nil {
		return "", err
	}

	hashSum := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashSum)

	return hashString, nil
}

// containsAny checks if a string contains any of the substrings.
func containsAny(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}

	return false
}
