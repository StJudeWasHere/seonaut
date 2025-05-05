package urlutils_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/stjudewashere/seonaut/internal/urlutils"
)

// Test AbsoluteURL with an html document that does not have a base tag.
func TestAbsoluteURLWithoutBase(t *testing.T) {
	urlStr := "https://example.com/"
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		t.Errorf("error parsing url: %v", err)
	}

	html := strings.NewReader(`
		<html>
			<head></head>
			<body></body>
		</html>`)

	doc, err := htmlquery.Parse(html)
	if err != nil {
		t.Errorf("error parsing html")
	}

	table := []struct {
		linkURL     string
		expectedURL string
	}{
		{"/test.html", "https://example.com/test.html"},
		{"test.html", "https://example.com/test.html"},
		{"/category/test.html", "https://example.com/category/test.html"},
		{"/category/../test.html", "https://example.com/test.html"},
		{"../test.html", "https://example.com/test.html"},
		{"https://example.com/test.html", "https://example.com/test.html"},
		{"https://external.com/test.html", "https://external.com/test.html"},
	}

	for _, u := range table {
		absolute, err := urlutils.AbsoluteURL(u.linkURL, doc, parsedURL)
		if err != nil {
			t.Errorf("absolute url error url %v", err)
		}

		if absolute.String() != u.expectedURL {
			t.Errorf("absolute url does not match expected value. Want %s Got %s", u.expectedURL, absolute)
		}
	}

}

// Test AbsoluteURL with a document that has a base tag.
func TestAbsoluteURLWithBase(t *testing.T) {
	urlStr := "https://example.com/"
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		t.Errorf("error parsing url: %v", err)
	}

	html := strings.NewReader(`
		<html>
			<head>
			<base href="https://example.com/base">
			</head>
			<body></body>
		</html>`)

	doc, err := htmlquery.Parse(html)
	if err != nil {
		t.Errorf("error parsing html")
	}

	table := []struct {
		linkURL     string
		expectedURL string
	}{
		{"/test.html", "https://example.com/base/test.html"},
		{"test.html", "https://example.com/base/test.html"},
		{"/category/test.html", "https://example.com/base/category/test.html"},
		{"/category/../test.html", "https://example.com/base/test.html"},
		{"../test.html", "https://example.com/test.html"},
		{"https://example.com/test.html", "https://example.com/test.html"},
		{"https://external.com/test.html", "https://external.com/test.html"},
	}

	for _, u := range table {
		absolute, err := urlutils.AbsoluteURL(u.linkURL, doc, parsedURL)
		if err != nil {
			t.Errorf("absolute url error url %v", err)
		}

		if absolute.String() != u.expectedURL {
			t.Errorf("absolute url does not match expected value. Want %s Got %s", u.expectedURL, absolute)
		}
	}
}
