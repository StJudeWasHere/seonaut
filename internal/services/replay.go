package services

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/stjudewashere/seonaut/internal/models"
	"golang.org/x/net/html"
)

type ReplayService struct{}

func NewReplayService() *ReplayService {
	return &ReplayService{}
}

// RewriteHTML rewrites the links and relevant URLs so they are handled by the proxy route.
func (r *ReplayService) RewriteHTML(htmlContent []byte, p *models.Project) ([]byte, error) {
	doc, err := htmlquery.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	projectURL, err := url.Parse(p.URL)
	if err != nil {
		return []byte{}, errors.New("error parsing projectURL")
	}

	rewriteAttr := func(node *html.Node, attrName string) {
		for i := range node.Attr {
			if node.Attr[i].Key == attrName {

				resolved, err := url.Parse(node.Attr[i].Val)
				if err != nil {
					continue
				}

				if !resolved.IsAbs() {
					resolved = projectURL.ResolveReference(resolved)
				}

				if resolved.Scheme != "http" && resolved.Scheme != "https" {
					continue
				}

				proxied := fmt.Sprintf("/replay?pid=%d&url=%s", p.Id, resolved.String())
				node.Attr[i].Val = proxied
			}
		}
	}

	for _, xpath := range []struct {
		xpath, attr string
	}{
		{`//img`, "src"},
		{`//script`, "src"},
		{`//link`, "href"},
		{`//iframe`, "src"},
		{`//source`, "src"},
		{`//video`, "poster"},
		{`//video`, "src"},
		{`//audio`, "src"},
		{`//a`, "href"},
	} {
		nodes := htmlquery.Find(doc, xpath.xpath)
		for _, node := range nodes {
			rewriteAttr(node, xpath.attr)
		}
	}

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// InjectHTML injects an HTML string into the HTML contents. It is used to inject a banner in
// all the html responses from the WACZ archive.
func (r *ReplayService) InjectHTML(htmlContent []byte, bannerHTML string) ([]byte, error) {
	doc, err := htmlquery.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	// Inject banner into <body>
	bodyNode := htmlquery.FindOne(doc, "//body")
	if bodyNode != nil {
		bannerFragment, err := html.ParseFragment(strings.NewReader(bannerHTML), bodyNode)
		if err != nil {
			return nil, err
		}

		// Insert each node before the existing first child to preserve order
		for i := len(bannerFragment) - 1; i >= 0; i-- {
			node := bannerFragment[i]
			node.Parent = bodyNode

			// Fix siblings
			node.NextSibling = bodyNode.FirstChild
			if bodyNode.FirstChild != nil {
				bodyNode.FirstChild.PrevSibling = node
			}
			bodyNode.FirstChild = node
		}
	}

	// Serialize the DOM back to HTML
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
