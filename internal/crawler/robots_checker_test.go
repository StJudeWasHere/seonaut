package crawler_test

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stjudewashere/seonaut/internal/crawler"
)

type MockClient struct{}

func (t *MockClient) Head(u string) (*crawler.ClientResponse, error) {
	return &crawler.ClientResponse{}, nil
}
func (t *MockClient) Get(u string) (*crawler.ClientResponse, error) {
	r := &http.Response{}
	if strings.HasPrefix(u, "https://example.com/") {
		body := `
		User-Agent: *
		Disallow: /disallowed
		Sitemap: /sitemap.xml
		`
		r.Body = io.NopCloser(bytes.NewBufferString(body))
		r.StatusCode = 200
	} else {
		r.Body = io.NopCloser(bytes.NewBufferString(""))
		r.StatusCode = 404
	}

	return &crawler.ClientResponse{Response: r}, nil
}

func (t *MockClient) GetUA() string {
	return "TEST UA"
}

// TestIsBlocked tests URLs allowed and disallowed in the robots.txt file.
func TestIsBlocked(t *testing.T) {
	robotsChecker := crawler.NewRobotsChecker(&MockClient{})
	u, err := url.Parse("https://example.com/disallowed")
	if err != nil {
		t.Errorf("url parse error %v", err)
	}

	if !robotsChecker.IsBlocked(u) {
		t.Errorf("Url %s should be blocked", u.String())
	}

	u, err = url.Parse("https://example.com/allowed")
	if err != nil {
		t.Errorf("url parse error %v", err)
	}

	if robotsChecker.IsBlocked(u) {
		t.Errorf("url %s should not be blocked", u.String())
	}
}

// TestRobotsExist tests if the robots.txt file exists for a given domain.
func TestRobotsExist(t *testing.T) {
	robotsChecker := crawler.NewRobotsChecker(&MockClient{})
	u, err := url.Parse("https://norobots.com/")
	if err != nil {
		t.Errorf("url parse error %v", err)
	}

	if robotsChecker.Exists(u) {
		t.Errorf("robots.txt should not exist in %s", u.String())
	}

	u, err = url.Parse("https://example.com/allowed")
	if err != nil {
		t.Errorf("url parse error %v", err)
	}

	if !robotsChecker.Exists(u) {
		t.Errorf("robots.txt should exist in %s", u.String())
	}
}

// TestGetSitemap tests if robots.txt file has a sitemaps list.
func TestGetSitemap(t *testing.T) {
	robotsChecker := crawler.NewRobotsChecker(&MockClient{})
	u, err := url.Parse("https://example.com/")
	if err != nil {
		t.Errorf("url parse error %v", err)
	}

	sitemaps := robotsChecker.GetSitemaps(u)
	if len(sitemaps) != 1 || sitemaps[0] != "/sitemap.xml" {
		t.Errorf("error getting sitemap from robots.txt in %s", u.String())
	}

	u, err = url.Parse("https://norobots.com/")
	if err != nil {
		t.Errorf("url parse error %v", err)
	}

	sitemaps = robotsChecker.GetSitemaps(u)
	if len(sitemaps) > 0 {
		t.Errorf("error getting sitemap from robots.txt in %s", u.String())
	}
}
