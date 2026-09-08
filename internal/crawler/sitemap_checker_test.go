package crawler_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stjudewashere/seonaut/internal/crawler"
)

// failingSitemapClient simulates a sitemap request that fails, which is what
// happens when a slow sitemap exceeds the HTTP client's timeout.
type failingSitemapClient struct{}

func (c *failingSitemapClient) Get(urlStr string) (*crawler.ClientResponse, error) {
	return nil, fmt.Errorf("Client.Timeout exceeded while awaiting headers")
}

func (c *failingSitemapClient) Head(urlStr string) (*crawler.ClientResponse, error) {
	return nil, fmt.Errorf("Client.Timeout exceeded while awaiting headers")
}

func (c *failingSitemapClient) GetUAName() string {
	return "TEST_UA"
}

// Test that ParseSitemaps returns even when the sitemap requests fail.
// Without the deferred wg.Done() the WaitGroup counter is never decremented
// and the crawler blocks forever before crawling a single URL.
func TestParseSitemapsReturnsOnRequestError(t *testing.T) {
	checker := crawler.NewSitemapChecker(&failingSitemapClient{}, 20000)

	done := make(chan struct{})
	go func() {
		defer close(done)
		checker.ParseSitemaps([]string{"https://example.com/sitemap.xml"}, func(u string) {})
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("expected ParseSitemaps to return, but it is still blocked")
	}
}

// Test that SitemapExists reports false when the request fails.
func TestSitemapExistsOnRequestError(t *testing.T) {
	checker := crawler.NewSitemapChecker(&failingSitemapClient{}, 20000)

	if checker.SitemapExists([]string{"https://example.com/sitemap.xml"}) {
		t.Fatal("expected SitemapExists to be false when the request fails")
	}
}
