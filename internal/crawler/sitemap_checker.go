package crawler

import (
	"net/http"

	"github.com/oxffaa/gopher-parse-sitemap"
)

type SitemapChecker struct{}

func NewSitemapChecker() *SitemapChecker {
	return &SitemapChecker{}
}

// Check if any of the sitemap URLs provided exist
func (sc *SitemapChecker) SitemapExists(URLs []string) bool {
	for _, s := range URLs {
		if sc.urlExists(s) == true {
			return true
		}
	}

	return false
}

// Check if a URL exists by checking its status code
func (sc *SitemapChecker) urlExists(URL string) bool {
	resp, err := http.Head(URL)
	if err != nil {
		return false
	}

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// Parse the sitemaps using a callback function on each entry
// For each URL provided check if it's an index sitemap
func (sc *SitemapChecker) ParseSitemaps(URLs []string, c func(u string)) {
	for _, l := range URLs {
		sitemaps := sc.checkIndex(l)
		for _, s := range sitemaps {
			sitemap.ParseFromSite(s, func(e sitemap.Entry) error {
				c(e.GetLocation())
				return nil
			})
		}
	}
}

// Returns a slice of strings with sitemap URLs
// If URL is a sitemap index the slice will contain all the sitemaps found
// Otherwise it will return an slice containing only the original URL
func (sc *SitemapChecker) checkIndex(URL string) []string {
	sitemaps := []string{}

	sitemap.ParseIndexFromSite(URL, func(e sitemap.IndexEntry) error {
		l := e.GetLocation()
		sitemaps = append(sitemaps, l)
		return nil
	})

	if len(sitemaps) == 0 {
		sitemaps = append(sitemaps, URL)
	}

	return sitemaps
}
