package crawler

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	// MaxBodySize is the limit of the retrieved response body in bytes.
	// The default value for MaxBodySize is 10MB (10 * 1024 * 1024 bytes).
	maxBodySize = 10 * 1024 * 1024

	// Number of threads a queue will use to crawl a project
	consumerThreads = 2
)

type Crawler struct {
	URL             *url.URL
	MaxPageReports  int
	IgnoreRobotsTxt bool
	FollowNofollow  bool
	IncludeNoindex  bool
	UserAgent       string
	CrawlSitemap    bool

	storage         *URLStorage
	sitemapChecker  *SitemapChecker
	sitemapExists   bool
	robotstxtExists bool
	plock           *sync.RWMutex
	responseCounter int
	robotsChecker   *RobotsChecker

	que            *que
	pr             chan<- PageReport
	client         *Client
	allowedDomains []string
}

func NewCrawler(url *url.URL, agent string, max int, irobots, fnofollow, inoindex, crawlSitemap bool) *Crawler {
	mainDomain := strings.TrimPrefix(url.Host, "www.")

	if url.Path == "" {
		url.Path = "/"
	}

	return &Crawler{
		URL:             url,
		MaxPageReports:  max,
		IgnoreRobotsTxt: irobots,
		FollowNofollow:  fnofollow,
		IncludeNoindex:  inoindex,
		UserAgent:       agent,
		CrawlSitemap:    crawlSitemap,

		storage:        NewURLStorage(),
		sitemapChecker: NewSitemapChecker(),
		plock:          &sync.RWMutex{},
		robotsChecker:  NewRobotsChecker(agent),

		que:            NewQueue(),
		client:         NewClient(agent),
		allowedDomains: []string{mainDomain, "www." + mainDomain},
	}
}

// Crawl starts crawling an URL and sends pagereports of the crawled URLs
// through the pr channel. It will end when there are no more URLs to crawl
// or the MaxPageReports limit is hit.
func (c *Crawler) Crawl(pr chan<- PageReport) {
	defer close(pr)

	c.pr = pr

	c.que.Push(c.URL.String())
	c.storage.Add(c.URL.String())

	sitemaps := c.getSitemaps()
	c.robotstxtExists = c.robotsChecker.Exists(c.URL)
	c.sitemapExists = c.sitemapChecker.SitemapExists(sitemaps)

	if c.CrawlSitemap && c.sitemapExists {
		go c.sitemapChecker.ParseSitemaps(sitemaps, c.loadSitemapURLs)
	}

	wg := new(sync.WaitGroup)
	wg.Add(consumerThreads)

	for i := 0; i < consumerThreads; i++ {
		go c.consumer(wg)
	}

	wg.Wait()
}

// Returns true if the sitemap.xml file exists
func (c *Crawler) SitemapExists() bool {
	return c.sitemapExists
}

// Returns true if the robots.txt file exists
func (c *Crawler) RobotstxtExists() bool {
	return c.robotstxtExists
}

// Consumer URLs from the queue
func (c *Crawler) consumer(w *sync.WaitGroup) {
	for {
		url, ok := c.que.Poll()
		if !ok {
			break
		}
		r, err := c.client.Get(url.(string))
		if err != nil {
			continue
		}
		c.responseHandler(r)
	}
	w.Done()
}

// Returns true if the page report limit has been hit
func (c *Crawler) isLimitHit() bool {
	c.plock.RLock()
	defer c.plock.RUnlock()

	return c.responseCounter >= c.MaxPageReports
}

// Increases the counter of crawled URLs
func (c *Crawler) increaseCrawledCounter() {
	c.plock.Lock()
	defer c.plock.Unlock()

	c.responseCounter++
}

// Returns true if the crawler is allowed to crawl the domain
func (c *Crawler) domainIsAllowed(s string) bool {
	for _, v := range c.allowedDomains {
		if v == s {
			return true
		}
	}

	return false
}

// Returns a list of sitemaps
func (c *Crawler) getSitemaps() []string {
	sm := c.robotsChecker.GetSitemaps(c.URL)
	if len(sm) > 0 {
		return sm
	}

	return []string{c.URL.Scheme + "://" + c.URL.Host + "/sitemap.xml"}
}

// Callback to load sitemap URLs into the queue
func (c *Crawler) loadSitemapURLs(u string) {
	if c.isLimitHit() {
		return
	}

	l, err := url.Parse(u)
	if err != nil {
		return
	}

	if l.Path == "/" {
		l.Path = "/"
	}

	c.que.Push(l.String())
}

// Handles the HTTP response
func (c *Crawler) responseHandler(r *http.Response) {
	defer func() {
		c.que.Ack(r.Request.URL.String())
		r.Body.Close()
	}()

	if c.isLimitHit() {
		return
	}

	var bodyReader io.Reader = r.Body
	bodyReader = io.LimitReader(bodyReader, int64(maxBodySize))

	b, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return
	}

	pageReport := NewPageReport(r.Request.URL, r.StatusCode, &r.Header, b)
	pageReport.BlockedByRobotstxt = c.robotsChecker.IsBlocked(r.Request.URL)

	if pageReport.Noindex == false || c.IncludeNoindex == true {
		pageReport.Crawled = true
		c.increaseCrawledCounter()
	}

	c.pr <- *pageReport

	if pageReport.Nofollow == true && c.FollowNofollow == false {
		return
	}

	for _, t := range c.getCrawlableURLs(pageReport) {
		if c.storage.Seen(t.String()) {
			continue
		}

		c.storage.Add(t.String())

		if c.IgnoreRobotsTxt == false && c.robotsChecker.IsBlocked(&t) {
			p := &PageReport{
				URL:                t.String(),
				ParsedURL:          &t,
				Crawled:            false,
				BlockedByRobotstxt: true,
			}

			c.pr <- *p

			continue
		}

		c.que.Push(t.String())
	}
}

// Returns all the crawlable URLs found in the document except the external URLs
func (c *Crawler) getCrawlableURLs(p *PageReport) []url.URL {
	var urls []url.URL
	var resources []string

	for _, l := range p.Links {
		if l.NoFollow && c.FollowNofollow == false {
			continue
		}

		if !c.domainIsAllowed(l.ParsedURL.Host) {
			continue
		}

		urls = append(urls, *l.ParsedURL)
	}

	for _, l := range p.Hreflangs {
		parsed, err := url.Parse(l.URL)
		if err != nil {
			continue
		}

		if !c.domainIsAllowed(parsed.Host) {
			continue
		}

		urls = append(urls, *parsed)
	}

	if p.RedirectURL != "" {
		parsed, err := url.Parse(p.RedirectURL)
		if err == nil && c.domainIsAllowed(parsed.Host) {
			urls = append(urls, *parsed)
		}
	}

	if p.Canonical != "" {
		parsed, err := url.Parse(p.Canonical)
		if err == nil && c.domainIsAllowed(parsed.Host) {
			urls = append(urls, *parsed)
		}
	}

	for _, l := range p.Scripts {
		resources = append(resources, l)
	}

	for _, l := range p.Styles {
		resources = append(resources, l)
	}

	for _, l := range p.Images {
		resources = append(resources, l.URL)
	}

	for _, v := range resources {
		t, err := url.Parse(v)
		if err != nil {
			continue
		}
		urls = append(urls, *t)
	}

	return urls
}
