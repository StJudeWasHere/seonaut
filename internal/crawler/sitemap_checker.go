package crawler

import (
	"errors"
	"sync"

	sitemap "github.com/oxffaa/gopher-parse-sitemap"
)

type SitemapChecker struct {
	limit  int
	client Client
}

func NewSitemapChecker(client Client, limit int) *SitemapChecker {
	return &SitemapChecker{
		limit:  limit,
		client: client,
	}
}

// Check if any of the sitemap URLs provided exist
func (sc *SitemapChecker) SitemapExists(URLs []string) bool {
	for _, s := range URLs {
		if sc.urlExists(s) {
			return true
		}
	}

	return false
}

// Check if a URL exists by checking its status code
func (sc *SitemapChecker) urlExists(URL string) bool {
	resp, err := sc.client.Head(URL)
	if err != nil {
		return false
	}

	return resp.Response.StatusCode >= 200 && resp.Response.StatusCode < 300
}

// Parse the sitemaps using a callback function on each entry
// For each URL provided check if it's an index sitemap
func (sc *SitemapChecker) ParseSitemaps(URLs []string, callback func(u string)) {
	c := 0
	wg := new(sync.WaitGroup)
	lock := sync.RWMutex{}

	for _, l := range URLs {
		sitemaps := sc.checkIndex(l)
		for _, s := range sitemaps {
			wg.Add(1)

			// Each sitemap is parsed in its own Go routine
			// If the sitemap limit is hit the parser function returns an error to stop the process
			go func(s string) {
				resp, err := sc.client.Get(s)
				if err != nil {
					return
				}
				defer resp.Response.Body.Close()

				sitemap.Parse(resp.Response.Body, func(e sitemap.Entry) error {
					callback(e.GetLocation())

					lock.Lock()
					defer lock.Unlock()

					c++
					if c >= sc.limit {
						return errors.New("URL limit hit")
					}

					return nil
				})

				wg.Done()
			}(s)
		}
	}

	wg.Wait()
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
