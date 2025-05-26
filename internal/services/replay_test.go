package services_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/stjudewashere/seonaut/internal/services"
)

var testReplayService = services.NewReplayService()

// Test if the URLs in the HTML are being rewritten.
func TestRewriteHTML(t *testing.T) {
	rewriteFunc := func(urlStr string) string {
		return fmt.Sprintf("/rewrite?url=%s", urlStr)
	}

	tests := []struct {
		name  string
		html  string
		xpath string
		attr  string
		want  string
	}{
		{
			name:  "Test link rewrite",
			html:  `<html><head><link href="styles.css"/></head></html>`,
			xpath: "//link",
			attr:  "href",
			want:  "/rewrite?url=styles.css",
		},
		{
			name:  "Test script rewrite",
			html:  `<html><head><script src="script.js"></script></head></html>`,
			xpath: "//script",
			attr:  "src",
			want:  "/rewrite?url=script.js",
		},
		{
			name:  "Test img rewrite",
			html:  `<html><body><img src="image.jpg"/></body></html>`,
			xpath: "//img",
			attr:  "src",
			want:  "/rewrite?url=image.jpg",
		},
		{
			name:  "Test anchor rewrite",
			html:  `<html><body><a href="page.html">Link</a></body></html>`,
			xpath: "//a",
			attr:  "href",
			want:  "/rewrite?url=page.html",
		},
		{
			name:  "Test iframe rewrite",
			html:  `<html><body><iframe src="page.html"></iframe></body></html>`,
			xpath: "//iframe",
			attr:  "src",
			want:  "/rewrite?url=page.html",
		},
		{
			name:  "Test audio rewrite",
			html:  `<html><body><audio src="page.wav"></audio></body></html>`,
			xpath: "//audio",
			attr:  "src",
			want:  "/rewrite?url=page.wav",
		},
		{
			name:  "Test vide rewrite",
			html:  `<html><body><video src="page.mp4"></vide></body></html>`,
			xpath: "//video",
			attr:  "src",
			want:  "/rewrite?url=page.mp4",
		},
		{
			name:  "Test video poster rewrite",
			html:  `<html><body><video src="page.mp4" poster="poster.jpg"></video></body></html>`,
			xpath: "//video",
			attr:  "poster",
			want:  "/rewrite?url=poster.jpg",
		},
		{
			name:  "Test source src rewrite",
			html:  `<html><body><<video><source src="page.mp4"></source></video></body></html>`,
			xpath: "//source",
			attr:  "src",
			want:  "/rewrite?url=page.mp4",
		},
		{
			name:  "Test inline style attribute rewrite",
			html:  `<html><body><div style="background-image: url('bg.jpg')"></div></body></html>`,
			xpath: "//div",
			attr:  "style",
			want:  "background-image: url(/rewrite?url=bg.jpg)",
		},
		{
			name:  "Test style tag url rewrite",
			html:  `<html><head><style>body { background: url('bg.jpg'); }</style></head></html>`,
			xpath: "//style",
			want:  "body { background: url(/rewrite?url=bg.jpg); }",
		},
		{
			name:  "Test image srcset rewrite",
			html:  `<html><body><img srcset="header320.png, header640.png 640w, header960.png 960w"></body></html>`,
			xpath: "//img",
			attr:  "srcset",
			want:  "/rewrite?url=header320.png, /rewrite?url=header640.png 640w, /rewrite?url=header960.png 960w",
		},
		{
			name:  "Test picture source srcset rewrite",
			html:  `<html><body><picture><source srcset="header320.png, header640.png 640w, header960.png 960w"></picture></body></html>`,
			xpath: "//picture/source",
			attr:  "srcset",
			want:  "/rewrite?url=header320.png, /rewrite?url=header640.png 640w, /rewrite?url=header960.png 960w",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rewrittenHTML, err := testReplayService.RewriteHTML([]byte(tt.html), rewriteFunc)
			if err != nil {
				t.Fatalf("RewriteHTML failed: %v", err)
			}

			doc, err := htmlquery.Parse(bytes.NewReader(rewrittenHTML))
			if err != nil {
				t.Fatalf("Failed to parse rewritten HTML: %v", err)
			}

			node := htmlquery.FindOne(doc, tt.xpath)
			if node == nil {
				t.Fatalf("Element %s not found", tt.xpath)
			}

			// Special case for style tag content as it rewrites the data in the tag instead
			// of rewritting the data in an attribute
			if tt.xpath == "//style" {
				if node.FirstChild == nil || node.FirstChild.Data != tt.want {
					t.Errorf("got %s, want %s", node.FirstChild.Data, tt.want)
				}
				return
			}

			if attrVal := htmlquery.SelectAttr(node, tt.attr); attrVal != tt.want {
				t.Errorf("got %s, want %s", attrVal, tt.want)
			}
		})
	}
}

// Test if the URLs in the css are being rewritten.
func TestRewriteCSS(t *testing.T) {
	rewriteFunc := func(urlStr string) string {
		return fmt.Sprintf("/rewrite?url=%s", urlStr)
	}

	css := `
	body { background: url('bg.jpg'); }
	div { background-image: url('div.jpg'); }
	`

	expected := `
	body { background: url(/rewrite?url=bg.jpg); }
	div { background-image: url(/rewrite?url=div.jpg); }
	`
	rewrittenCSS := testReplayService.RewriteCSS(css, rewriteFunc)
	if rewrittenCSS != expected {
		t.Errorf("got %s want %s", rewrittenCSS, expected)
	}
}

// Test the HTML injection method.
func TestInjectHTML(t *testing.T) {
	html := `<html><head></head><body></body></html>`
	script := `<script>console.log("test")</script>`
	banner := `<div id="banner">Test Banner</div>`

	rewrittenHTML, err := testReplayService.InjectHTML([]byte(html), script, banner)
	if err != nil {
		t.Fatalf("InjectHTML failed: %v", err)
	}

	doc, err := htmlquery.Parse(bytes.NewReader(rewrittenHTML))
	if err != nil {
		t.Fatalf("Failed to parse rewritten HTML: %v", err)
	}

	// Check if script is in head
	scriptNode := htmlquery.FindOne(doc, "//head/script")
	if scriptNode == nil {
		t.Error("Script was not injected in head")
	}

	// Check if banner is in body
	bannerNode := htmlquery.FindOne(doc, "//body/div[@id='banner']")
	if bannerNode == nil {
		t.Error("Banner was not injected in body")
	}
}
