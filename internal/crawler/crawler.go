package crawler

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/stjudewashere/seonaut/internal/html_parser"
	"github.com/stjudewashere/seonaut/internal/httpcrawler"
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/queue"
	"github.com/stjudewashere/seonaut/internal/urlstorage"
)

type Options struct {
	MaxPageReports     int
	IgnoreRobotsTxt    bool
	FollowNofollow     bool
	IncludeNoindex     bool
	UserAgent          string
	CrawlSitemap       bool
	AllowSubdomains    bool
	BasicAuth          bool
	AuthUser           string
	AuthPass           string
	CheckExternalLinks bool
}

type Client interface {
	Get(u string) (*http.Response, error)
	Head(u string) (*http.Response, error)
	Do(req *http.Request) (*http.Response, error)
}
type Crawler struct {
	url                 *url.URL
	options             *Options
	queue               *queue.Queue
	storage             *urlstorage.URLStorage
	sitemapStorage      *urlstorage.URLStorage
	sitemapChecker      *httpcrawler.SitemapChecker
	sitemapExists       bool
	sitemapIsBlocked    bool
	sitemaps            []string
	robotstxtExists     bool
	responseCounter     int
	robotsChecker       *httpcrawler.RobotsChecker
	prStream            chan *models.PageReportMessage
	allowedDomains      map[string]bool
	mainDomain          string
	httpCrawler         *httpcrawler.HttpCrawler
	qStream             chan *httpcrawler.RequestMessage
	cancel              context.CancelFunc
	crawling            bool
	crawlLock           sync.RWMutex
	client              Client
	externalLinksStatus map[string]int
}

func NewCrawler(url *url.URL, options *Options) *Crawler {
	mainDomain := strings.TrimPrefix(url.Host, "www.")

	if url.Path == "" {
		url.Path = "/"
	}

	storage := urlstorage.New()
	storage.Add(url.String())

	httpClient := httpcrawler.NewClient(&httpcrawler.ClientOptions{
		UserAgent:        options.UserAgent,
		BasicAuth:        options.BasicAuth,
		BasicAuthDomains: []string{mainDomain, "www." + mainDomain},
		AuthUser:         options.AuthUser,
		AuthPass:         options.AuthPass,
	})

	qctx, cancelQueue := context.WithCancel(context.Background())
	q := queue.New(qctx)

	robotsChecker := httpcrawler.NewRobotsChecker(httpClient, options.UserAgent)
	if !robotsChecker.IsBlocked(url) || options.IgnoreRobotsTxt {
		q.Push(&httpcrawler.RequestMessage{URL: url.String()})
	}

	sitemaps := robotsChecker.GetSitemaps(url)
	nonBlockedSitemaps := []string{}
	if len(sitemaps) == 0 {
		sitemaps = []string{url.Scheme + "://" + url.Host + "/sitemap.xml"}
	}

	sitemapIsBlocked := false
	for _, sm := range sitemaps {
		parsedSm, err := url.Parse(sm)
		if err != nil {
			continue
		}

		if robotsChecker.IsBlocked(parsedSm) {
			sitemapIsBlocked = true
			if !options.IgnoreRobotsTxt {
				continue
			}
		}

		nonBlockedSitemaps = append(nonBlockedSitemaps, sm)
	}

	sitemaps = nonBlockedSitemaps

	sitemapChecker := httpcrawler.NewSitemapChecker(httpClient, options.MaxPageReports)
	qStream := make(chan *httpcrawler.RequestMessage)
	ctx, cancel := context.WithCancel(context.Background())

	c := &Crawler{
		url:                 url,
		options:             options,
		queue:               q,
		storage:             storage,
		sitemapStorage:      urlstorage.New(),
		sitemapChecker:      sitemapChecker,
		sitemapExists:       sitemapChecker.SitemapExists(sitemaps),
		sitemapIsBlocked:    sitemapIsBlocked,
		sitemaps:            sitemaps,
		robotsChecker:       robotsChecker,
		robotstxtExists:     robotsChecker.Exists(url),
		allowedDomains:      map[string]bool{mainDomain: true, "www." + mainDomain: true},
		mainDomain:          mainDomain,
		prStream:            make(chan *models.PageReportMessage),
		qStream:             qStream,
		httpCrawler:         httpcrawler.New(httpClient, qStream),
		cancel:              cancel,
		crawling:            true,
		client:              httpClient,
		externalLinksStatus: make(map[string]int),
	}

	go c.queueStreamer(qctx)
	go func() {
		defer func() {
			cancelQueue()
			cancel()
		}()

		c.crawl(ctx)
	}()

	return c
}

// Returns the PageReportMessage channel that streams all generated PageReports
// into a PageReportMessage struct.
func (c *Crawler) Stream() <-chan *models.PageReportMessage {
	return c.prStream
}

// Polls URLs from the queue and sends them into the qStream channel.
// queueStreamer shuts down when the ctx context is done.
func (c *Crawler) queueStreamer(ctx context.Context) {
	defer close(c.qStream)
	for {
		select {
		case <-ctx.Done():
			return
		case c.qStream <- c.queue.Poll():
		}
	}
}

// Crawl starts crawling an URL and sends pagereports of the crawled URLs
// through the pr channel. It will end when there are no more URLs to crawl
// or the MaxPageReports limit is hit.
func (c *Crawler) crawl(ctx context.Context) {
	defer close(c.prStream)

	if c.sitemapExists && c.options.CrawlSitemap {
		c.sitemapChecker.ParseSitemaps(c.sitemaps, c.loadSitemapURLs)
	}

	sitemapLoaded := false
	if !c.queue.Active() && c.options.CrawlSitemap {
		c.queueSitemapURLs()
		sitemapLoaded = true
	}

	if !c.queue.Active() {
		return
	}

	for rm := range c.httpCrawler.Crawl(ctx) {
		err := c.handleResponse(rm)
		if err != nil {
			log.Printf("handleResponse %s: Error %v", rm.URL, err)
		}

		if !c.queue.Active() && c.options.CrawlSitemap && !sitemapLoaded {
			c.queueSitemapURLs()
			sitemapLoaded = true
		}

		if !c.queue.Active() || c.responseCounter >= c.options.MaxPageReports {
			break
		}
	}
}

// handleResponse handles the crawler response messages.
// It creates a new PageReport and adds the new URLs to the crawler queue.
func (c *Crawler) handleResponse(r *httpcrawler.ResponseMessage) error {
	c.queue.Ack(r.URL)
	if r.Error != nil {
		return r.Error
	}

	defer r.Response.Body.Close()

	pageReport, htmlNode, err := html_parser.NewFromHTTPResponse(r.Response)
	if err != nil {
		return err
	}

	parsedURL, err := url.Parse(r.URL)
	if err != nil {
		return err
	}

	pageReport.Depth = r.Depth
	pageReport.BlockedByRobotstxt = c.robotsChecker.IsBlocked(parsedURL)
	pageReport.InSitemap = c.sitemapStorage.Seen(r.URL)
	pageReport.Crawled = true

	c.responseCounter++

	urls := []*url.URL{}
	if c.options.FollowNofollow || !pageReport.Nofollow {
		crawlable := [][]*url.URL{
			c.getCrawlableLinks(pageReport),
			c.getResourceURLs(pageReport),
			c.getCrawlableURLs(pageReport),
		}

		for _, c := range crawlable {
			urls = append(urls, c...)
		}
	}

	c.crawlLock.RLock()
	crawling := c.crawling
	c.crawlLock.RUnlock()

	for _, t := range urls {
		if c.storage.Seen(t.String()) {
			continue
		}

		c.storage.Add(t.String())

		if !c.options.IgnoreRobotsTxt && c.robotsChecker.IsBlocked(t) {
			c.prStream <- &models.PageReportMessage{
				Crawled:    c.responseCounter,
				Crawling:   crawling,
				Discovered: c.queue.Count(),
				HtmlNode:   htmlNode,
				Header:     &r.Response.Header,
				PageReport: &models.PageReport{
					URL:                t.String(),
					ParsedURL:          t,
					Crawled:            false,
					BlockedByRobotstxt: true,
				},
			}

			continue
		}

		c.queue.Push(&httpcrawler.RequestMessage{URL: t.String(), Depth: pageReport.Depth + 1})
	}

	if c.options.CheckExternalLinks {
		// Check external links status code
		for i, e := range pageReport.ExternalLinks {
			status, ok := c.externalLinksStatus[e.URL]
			if ok {
				pageReport.ExternalLinks[i].StatusCode = status
				continue
			}

			resp, err := c.client.Head(e.URL)
			if err != nil {
				continue
			}
			c.externalLinksStatus[e.URL] = resp.StatusCode
			pageReport.ExternalLinks[i].StatusCode = resp.StatusCode
		}
	}

	if !pageReport.Noindex || c.options.IncludeNoindex {
		c.prStream <- &models.PageReportMessage{
			PageReport: pageReport,
			HtmlNode:   htmlNode,
			Header:     &r.Response.Header,
			Crawled:    c.responseCounter,
			Crawling:   crawling,
			Discovered: c.queue.Count(),
		}
	}

	return nil
}

// Returns true if the crawler is allowed to crawl the domain, checking the allowedDomains slice.
// If the AllowSubdomains option is set, returns true the given domain is a subdomain of the
// crawlers's base domain.
func (c *Crawler) domainIsAllowed(s string) bool {
	_, ok := c.allowedDomains[s]
	if ok {
		return true
	}

	if c.options.AllowSubdomains && strings.HasSuffix(s, c.mainDomain) {
		return true
	}

	return false
}

// Callback to load sitemap URLs into the sitemap storage
func (c *Crawler) loadSitemapURLs(u string) {
	l, err := url.Parse(u)
	if err != nil {
		return
	}

	if l.Path == "" {
		l.Path = "/"
	}

	c.sitemapStorage.Add(l.String())
}

// queueSitemapURLs loops through the sitemap's URLs, adding any unseen URLs to the crawler's queue.
func (c *Crawler) queueSitemapURLs() {
	c.sitemapStorage.Iterate(func(v string) {
		if !c.storage.Seen(v) {
			c.storage.Add(v)
			c.queue.Push(&httpcrawler.RequestMessage{URL: v})
		}
	})
}

// Returns true if the sitemap.xml file exists
func (c *Crawler) SitemapExists() bool {
	return c.sitemapExists
}

// Returns true if the robots.txt file exists
func (c *Crawler) RobotstxtExists() bool {
	return c.robotstxtExists
}

// Returns true if any of the website's sitemaps is blocked in the robots.txt file
func (c *Crawler) SitemapIsBlocked() bool {
	return c.sitemapIsBlocked
}

// Returns a slice with all the crawlable Links from the PageReport's links.
// URLs extracted from internal Links and ExternalLinks are crawlable only if the domain name is allowed and
// if they don't have the "nofollow" attribute. If they have the "nofollow" attribute, they are also considered
// crawlable if the crawler's FollowNofollow option is enabled.
func (c *Crawler) getCrawlableLinks(p *models.PageReport) []*url.URL {
	var urls []*url.URL

	links := append(p.Links, p.ExternalLinks...)
	for _, l := range links {
		if l.ParsedURL.Host == c.url.Host && !strings.HasPrefix(l.ParsedURL.Path, c.url.Path) {
			continue
		}

		if !c.domainIsAllowed(l.ParsedURL.Host) {
			continue
		}

		if l.NoFollow && !c.options.FollowNofollow {
			continue
		}

		urls = append(urls, l.ParsedURL)
	}

	return urls
}

// Returns a slice containing all the resource URLs from a PageReport.
// The resource URLs are always considered crawlable.
func (c *Crawler) getResourceURLs(p *models.PageReport) []*url.URL {
	var urls []*url.URL
	var resources []string

	for _, l := range p.Images {
		resources = append(resources, l.URL)
	}

	resources = append(resources, p.Scripts...)
	resources = append(resources, p.Styles...)
	resources = append(resources, p.Audios...)
	resources = append(resources, p.Videos...)

	for _, v := range resources {
		t, err := url.Parse(v)
		if err != nil {
			continue
		}
		urls = append(urls, t)
	}

	return urls
}

// Returns a slice of crawlable URLs extracted from the Hreflangs, Iframes,
// Redirect URLs and Canonical URLs found in the PageReport.
// The URLs are considered crawlable only if its domain is allowed by the crawler.
func (c *Crawler) getCrawlableURLs(p *models.PageReport) []*url.URL {
	var urls []*url.URL
	var resources []string

	for _, l := range p.Hreflangs {
		resources = append(resources, l.URL)
	}

	resources = append(resources, p.Iframes...)

	if p.RedirectURL != "" {
		resources = append(resources, p.RedirectURL)
	}

	if p.Canonical != "" {
		resources = append(resources, p.Canonical)
	}

	for _, r := range resources {
		parsed, err := url.Parse(r)
		if err != nil {
			continue
		}

		if c.domainIsAllowed(parsed.Host) {
			urls = append(urls, parsed)
		}
	}

	return urls
}

// Stops the cralwer by canceling the cralwer context.
func (c *Crawler) Stop() {
	c.cancel()
	c.crawlLock.Lock()
	c.crawling = false
	c.crawlLock.Unlock()
}
