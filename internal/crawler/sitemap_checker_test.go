package crawler_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stjudewashere/seonaut/internal/crawler"
)

const sitemapTestTimeout = 5 * time.Second

// sitemapMockClient serves a valid urlset for any URL containing "good" and
// returns an error for any URL containing "bad", simulating a sitemap that is
// unreachable at request time.
type sitemapMockClient struct {
	mu       sync.Mutex
	requests []string
}

func (m *sitemapMockClient) Head(u string) (*crawler.ClientResponse, error) {
	return &crawler.ClientResponse{Response: &http.Response{StatusCode: http.StatusOK}}, nil
}

func (m *sitemapMockClient) Get(u string) (*crawler.ClientResponse, error) {
	m.mu.Lock()
	m.requests = append(m.requests, u)
	m.mu.Unlock()

	if strings.Contains(u, "bad") {
		return nil, errors.New("mock transport error")
	}

	body := `<?xml version="1.0" encoding="UTF-8"?>
	<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
		<url><loc>https://example.com/page1</loc></url>
		<url><loc>https://example.com/page2</loc></url>
	</urlset>`

	return &crawler.ClientResponse{
		Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		},
	}, nil
}

func (m *sitemapMockClient) GetUAName() string {
	return "SEOnautBot"
}

// newSitemapIndexServer returns a test server serving a sitemap index that
// references the given sitemap locations. ParseSitemaps resolves the index over
// the network, so a local server keeps the test hermetic.
func newSitemapIndexServer(t *testing.T, locs ...string) *httptest.Server {
	t.Helper()

	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString(`<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	for _, l := range locs {
		sb.WriteString("<sitemap><loc>" + l + "</loc></sitemap>")
	}
	sb.WriteString(`</sitemapindex>`)
	index := sb.String()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = io.WriteString(w, index)
	}))
	t.Cleanup(server.Close)

	return server
}

// runParseSitemaps runs ParseSitemaps with a timeout, collecting the URLs passed
// to the callback. It fails the test rather than hanging if ParseSitemaps never
// returns, so a regression shows up as a failure instead of a stuck CI job.
func runParseSitemaps(t *testing.T, sc *crawler.SitemapChecker, urls []string) []string {
	t.Helper()

	var (
		mu   sync.Mutex
		seen []string
	)

	done := make(chan struct{})
	go func() {
		defer close(done)
		sc.ParseSitemaps(urls, func(u string) {
			mu.Lock()
			defer mu.Unlock()
			seen = append(seen, u)
		})
	}()

	select {
	case <-done:
	case <-time.After(sitemapTestTimeout):
		t.Fatalf("ParseSitemaps did not return within %s; the WaitGroup was never released", sitemapTestTimeout)
	}

	mu.Lock()
	defer mu.Unlock()

	return append([]string(nil), seen...)
}

// TestParseSitemapsReturnsWhenRequestFails makes sure ParseSitemaps returns when
// every sitemap request fails. Each sitemap goroutine has to decrement the
// WaitGroup even on the error path, otherwise Wait blocks forever and the crawl
// never starts.
func TestParseSitemapsReturnsWhenRequestFails(t *testing.T) {
	server := newSitemapIndexServer(t, "https://example.com/bad1.xml", "https://example.com/bad2.xml")

	sc := crawler.NewSitemapChecker(&sitemapMockClient{}, 100)

	if urls := runParseSitemaps(t, sc, []string{server.URL}); len(urls) != 0 {
		t.Errorf("expected no urls from failing sitemaps, got %d", len(urls))
	}
}

// TestParseSitemapsWithFailingAndWorkingSitemaps makes sure one unreachable
// sitemap does not prevent the remaining sitemaps from being parsed. A sitemap
// index commonly fans out to several children, and a single failed request must
// not discard the others.
func TestParseSitemapsWithFailingAndWorkingSitemaps(t *testing.T) {
	server := newSitemapIndexServer(t, "https://example.com/bad.xml", "https://example.com/good.xml")

	sc := crawler.NewSitemapChecker(&sitemapMockClient{}, 100)

	urls := runParseSitemaps(t, sc, []string{server.URL})
	if len(urls) != 2 {
		t.Fatalf("expected 2 urls from the working sitemap, got %d: %v", len(urls), urls)
	}

	for _, want := range []string{"https://example.com/page1", "https://example.com/page2"} {
		found := false
		for _, got := range urls {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected url %s in %v", want, urls)
		}
	}
}

// TestParseSitemapsRespectsLimit makes sure the callback is not called more than
// the configured limit allows.
func TestParseSitemapsRespectsLimit(t *testing.T) {
	server := newSitemapIndexServer(t, "https://example.com/good.xml")

	sc := crawler.NewSitemapChecker(&sitemapMockClient{}, 1)

	if urls := runParseSitemaps(t, sc, []string{server.URL}); len(urls) > 1 {
		t.Errorf("expected at most 1 url with a limit of 1, got %d: %v", len(urls), urls)
	}
}
