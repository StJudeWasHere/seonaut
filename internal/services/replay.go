package services

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type rewriteURL func(string) string

type ReplayService struct{}

func NewReplayService() *ReplayService {
	return &ReplayService{}
}

// RewriteHTML rewrites the links and relevant URLs so they are handled by the proxy route.
func (r *ReplayService) RewriteHTML(htmlContent []byte, rewriteFunc rewriteURL) ([]byte, error) {
	doc, err := htmlquery.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	// Rewrite the URLs in the specified element and attribute.
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
			for i := range node.Attr {
				if node.Attr[i].Key == xpath.attr {
					node.Attr[i].Val = rewriteFunc(node.Attr[i].Val)
				}
			}
		}
	}

	// Rewrite urls in style blocks
	styleElements := htmlquery.Find(doc, `//style`)
	for _, el := range styleElements {
		if el.FirstChild != nil && el.FirstChild.Type == html.TextNode {
			el.FirstChild.Data = r.RewriteCSS(el.FirstChild.Data, rewriteFunc)
		}
	}

	// Rewrite urls in inline css
	// htmlquery panics if htmlNode is of type html.ErroNode
	if doc.Type != html.ErrorNode {
		inlineStyleElements, err := htmlquery.QueryAll(doc, "//*[@style]")
		if err != nil {
			log.Printf("error getting elements with style attribute: %v", err)
		}

		for _, inlineStyleElement := range inlineStyleElements {
			for i, attr := range inlineStyleElement.Attr {
				if attr.Key == "style" {
					inlineStyleElement.Attr[i].Val = r.RewriteCSS(attr.Val, rewriteFunc)
				}
			}
		}
	}

	// Rewrite URLs in srcset tags of image and picture source elements.
	// The srcset can have multiple URLs with optional descriptors.
	// ex. <img src="header640.png 640w, header960.png 960w">
	srcsets, err := htmlquery.QueryAll(doc, "//img[@srcset] | //picture/source[@srcset]")
	if err != nil {
		log.Printf("error getting all srcset elements %v", err)
	}

	for _, srcset := range srcsets {
		rewrittenParts := []string{}
		urlset := htmlquery.SelectAttr(srcset, "srcset")
		urlset = strings.Trim(urlset, ",")
		parts := strings.Split(urlset, ", ")

		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}

			pp := strings.SplitN(p, " ", 2)
			rewrittenURL := rewriteFunc(pp[0])
			if len(pp) > 1 {
				rewrittenURL += " " + pp[1]
			}
			rewrittenParts = append(rewrittenParts, rewrittenURL)
		}

		for i := range srcset.Attr {
			if srcset.Attr[i].Key == "srcset" {
				srcset.Attr[i].Val = strings.Join(rewrittenParts, ", ")
			}
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
func (r *ReplayService) InjectHTML(htmlContent []byte, scripts string, bannerHTML string) ([]byte, error) {
	doc, err := htmlquery.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	// Inject banner into <head>
	headNode := htmlquery.FindOne(doc, "//head")
	if headNode != nil {
		bannerFragment, err := html.ParseFragment(strings.NewReader(scripts), headNode)
		if err != nil {
			return nil, err
		}

		// Insert each node before the existing first child to preserve order
		for i := len(bannerFragment) - 1; i >= 0; i-- {
			node := bannerFragment[i]
			node.Parent = headNode

			// Fix siblings
			node.NextSibling = headNode.FirstChild
			if headNode.FirstChild != nil {
				headNode.FirstChild.PrevSibling = node
			}
			headNode.FirstChild = node
		}
	}

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

// RewriteCSSURLs parses css content and rewrites the urls in URLTokens so they are
// handled by the proxy route.
func (r *ReplayService) RewriteCSS(cssContent string, rewriteFunc rewriteURL) string {
	urlRegex := regexp.MustCompile(`url\((.*?)\)`)
	rewrittenCSS := urlRegex.ReplaceAllStringFunc(cssContent, func(match string) string {
		urlStr := strings.TrimPrefix(strings.TrimSuffix(match, ")"), "url(")
		urlStr = strings.Trim(urlStr, "'\"") // Remove any surrounding quotes
		newURL := rewriteFunc(urlStr)
		return fmt.Sprintf("url(%s)", newURL)
	})

	return rewrittenCSS
}
