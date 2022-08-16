package crawler

import (
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	// MaxBodySize is the limit of the retrieved response body in bytes.
	// The default value for MaxBodySize is 10MB (10 * 1024 * 1024 bytes).
	maxBodySize = 10 * 1024 * 1024

	// Random delay in milliseconds.
	// A random delay up to this value is introduced before new HTTP requests.
	randomDelay = 1500

	// Number of threads a queue will use to crawl a project.
	consumerThreads = 2
)

type Options struct {
	MaxPageReports  int
	IgnoreRobotsTxt bool
	FollowNofollow  bool
	IncludeNoindex  bool
	UserAgent       string
	CrawlSitemap    bool
	AllowSubdomains bool
}

type Crawler struct {
	URL     *url.URL
	Options *Options

	storage         *URLStorage
	sitemapChecker  *SitemapChecker
	sitemapExists   bool
	robotstxtExists bool
	plock           *sync.RWMutex
	responseCounter int
	robotsChecker   *RobotsChecker

	queue          *queue
	prStream       chan *PageReportMessage
	client         *Client
	allowedDomains []string
}

func NewCrawler(url *url.URL, options *Options) *Crawler {
	mainDomain := strings.TrimPrefix(url.Host, "www.")

	if url.Path == "" {
		url.Path = "/"
	}

	return &Crawler{
		URL:     url,
		Options: options,

		storage:        NewURLStorage(),
		sitemapChecker: NewSitemapChecker(),
		plock:          &sync.RWMutex{},
		robotsChecker:  NewRobotsChecker(options.UserAgent),

		queue:          NewQueue(),
		client:         NewClient(options.UserAgent),
		allowedDomains: []string{mainDomain, "www." + mainDomain},
		prStream:       make(chan *PageReportMessage),
	}
}

// Crawl starts crawling an URL and sends pagereports of the crawled URLs
// through the pr channel. It will end when there are no more URLs to crawl
// or the MaxPageReports limit is hit.
func (c *Crawler) Crawl() <-chan *PageReportMessage {
	go func() {
		defer close(c.prStream)

		c.queue.Push(c.URL.String())
		c.storage.Add(c.URL.String())

		sitemaps := c.getSitemaps()
		c.robotstxtExists = c.robotsChecker.Exists(c.URL)
		c.sitemapExists = c.sitemapChecker.SitemapExists(sitemaps)

		if c.Options.CrawlSitemap && c.sitemapExists {
			go c.sitemapChecker.ParseSitemaps(sitemaps, c.loadSitemapURLs)
		}

		wg := new(sync.WaitGroup)
		wg.Add(consumerThreads)

		for i := 0; i < consumerThreads; i++ {
			go c.consumer(wg)
		}

		wg.Wait()
	}()

	return c.prStream
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
		if c.isLimitHit() {
			break
		}

		url, ok := c.queue.Poll()
		if !ok {
			break
		}

		time.Sleep(time.Duration(rand.Intn(randomDelay)) * time.Millisecond)

		r, err := c.client.Get(url.(string))
		if err != nil {
			c.queue.Ack(url.(string))
			continue
		}
		c.responseHandler(r)
		c.queue.Ack(url.(string))
	}
	w.Done()
}

// Returns true if the page report limit has been hit
func (c *Crawler) isLimitHit() bool {
	c.plock.RLock()
	defer c.plock.RUnlock()

	return c.responseCounter >= c.Options.MaxPageReports
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

	if c.Options.AllowSubdomains {
		return strings.HasSuffix(s, c.URL.Host)
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

	if c.storage.Seen(l.String()) {
		return
	}

	c.storage.Add(l.String())
	c.queue.Push(l.String())
}

// Handles the HTTP response
func (c *Crawler) responseHandler(r *http.Response) {
	defer r.Body.Close()

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

	if pageReport.Noindex == false || c.Options.IncludeNoindex == true {
		pageReport.Crawled = true
		c.increaseCrawledCounter()
	}

	crawled := c.getCrawledCount()
	discovered := c.getDiscoveredURLs()

	if pageReport.Nofollow == true && c.Options.FollowNofollow == false {
		c.prStream <- &PageReportMessage{
			PageReport: pageReport,
			Crawled:    crawled,
			Discovered: discovered,
		}

		return
	}

	for _, t := range c.getCrawlableURLs(pageReport) {
		if c.storage.Seen(t.String()) {
			continue
		}

		c.storage.Add(t.String())

		if c.Options.IgnoreRobotsTxt == false && c.robotsChecker.IsBlocked(t) {
			message := &PageReportMessage{
				Crawled:    crawled,
				Discovered: discovered,
			}

			message.PageReport = &PageReport{
				URL:                t.String(),
				ParsedURL:          t,
				Crawled:            false,
				BlockedByRobotstxt: true,
			}

			c.prStream <- message

			continue
		}

		c.queue.Push(t.String())
	}

	message := &PageReportMessage{
		PageReport: pageReport,
		Crawled:    crawled,
		Discovered: discovered,
	}

	c.prStream <- message
}

// Returns all the crawlable URLs found in the HTML document
func (c *Crawler) getCrawlableURLs(p *PageReport) []*url.URL {
	var urls []*url.URL
	var resources []string

	for _, l := range p.Links {
		if l.NoFollow && c.Options.FollowNofollow == false {
			continue
		}

		if !c.domainIsAllowed(l.ParsedURL.Host) {
			continue
		}

		urls = append(urls, l.ParsedURL)
	}

	for _, l := range p.ExternalLinks {
		if l.NoFollow && c.Options.FollowNofollow == false {
			continue
		}

		if !c.domainIsAllowed(l.ParsedURL.Host) {
			continue
		}

		urls = append(urls, l.ParsedURL)
	}

	for _, l := range p.Hreflangs {
		parsed, err := url.Parse(l.URL)
		if err != nil {
			continue
		}

		if !c.domainIsAllowed(parsed.Host) {
			continue
		}

		urls = append(urls, parsed)
	}

	for _, l := range p.Iframes {
		parsed, err := url.Parse(l)
		if err != nil {
			continue
		}

		if !c.domainIsAllowed(parsed.Host) {
			continue
		}

		urls = append(urls, parsed)
	}

	if p.RedirectURL != "" {
		parsed, err := url.Parse(p.RedirectURL)
		if err == nil && c.domainIsAllowed(parsed.Host) {
			urls = append(urls, parsed)
		}
	}

	if p.Canonical != "" {
		parsed, err := url.Parse(p.Canonical)
		if err == nil && c.domainIsAllowed(parsed.Host) {
			urls = append(urls, parsed)
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

	for _, l := range p.Audios {
		resources = append(resources, l)
	}

	for _, l := range p.Videos {
		resources = append(resources, l)
	}

	for _, v := range resources {
		t, err := url.Parse(v)
		if err != nil {
			continue
		}
		urls = append(urls, t)
	}

	return urls
}

// Returns the number of crawled URLs
func (c *Crawler) getCrawledCount() int {
	c.plock.Lock()
	l := c.responseCounter
	defer c.plock.Unlock()

	return l
}

// Returns the number of URLs currently in the queue
func (c *Crawler) getDiscoveredURLs() int {
	return c.queue.Count()
}
