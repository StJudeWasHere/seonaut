package crawler

import (
	"context"
	"log"
	"net/url"
	"strings"

	"github.com/stjudewashere/seonaut/internal/html_parser"
	"github.com/stjudewashere/seonaut/internal/http_crawler"
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/queue"
	"github.com/stjudewashere/seonaut/internal/urlstorage"
)

type Options struct {
	MaxPageReports  int
	IgnoreRobotsTxt bool
	FollowNofollow  bool
	IncludeNoindex  bool
	UserAgent       string
	CrawlSitemap    bool
	AllowSubdomains bool
	BasicAuth       bool
	AuthUser        string
	AuthPass        string
}

type Crawler struct {
	url             *url.URL
	options         *Options
	queue           *queue.Queue
	storage         *urlstorage.URLStorage
	sitemapStorage  *urlstorage.URLStorage
	sitemapChecker  *http_crawler.SitemapChecker
	sitemapExists   bool
	sitemaps        []string
	robotstxtExists bool
	responseCounter int
	robotsChecker   *http_crawler.RobotsChecker
	prStream        chan *models.PageReportMessage
	allowedDomains  map[string]bool
	mainDomain      string
	httpCrawler     *http_crawler.HttpCrawler
	qStream         chan string
}

func NewCrawler(url *url.URL, options *Options) *Crawler {
	mainDomain := strings.TrimPrefix(url.Host, "www.")

	if url.Path == "" {
		url.Path = "/"
	}

	storage := urlstorage.New()
	storage.Add(url.String())

	ctx, cancel := context.WithCancel(context.Background())

	q := queue.New(ctx)
	q.Push(url.String())

	httpClient := http_crawler.NewClient(&http_crawler.ClientOptions{
		UserAgent:        options.UserAgent,
		BasicAuth:        options.BasicAuth,
		BasicAuthDomains: []string{mainDomain, "www." + mainDomain},
		AuthUser:         options.AuthUser,
		AuthPass:         options.AuthPass,
	})

	robotsChecker := http_crawler.NewRobotsChecker(httpClient, options.UserAgent)

	sitemaps := robotsChecker.GetSitemaps(url)
	if len(sitemaps) == 0 {
		sitemaps = []string{url.Scheme + "://" + url.Host + "/sitemap.xml"}
	}

	sitemapChecker := http_crawler.NewSitemapChecker(httpClient, options.MaxPageReports)
	qStream := make(chan string)

	c := &Crawler{
		url:             url,
		options:         options,
		queue:           q,
		storage:         storage,
		sitemapStorage:  urlstorage.New(),
		sitemapChecker:  sitemapChecker,
		sitemapExists:   sitemapChecker.SitemapExists(sitemaps),
		sitemaps:        sitemaps,
		robotsChecker:   robotsChecker,
		robotstxtExists: robotsChecker.Exists(url),
		allowedDomains:  map[string]bool{mainDomain: true, "www." + mainDomain: true},
		mainDomain:      mainDomain,
		prStream:        make(chan *models.PageReportMessage),
		qStream:         qStream,
		httpCrawler:     http_crawler.New(httpClient, qStream),
	}

	go c.queueStreamer(ctx)
	go func() {
		c.crawl(ctx)
		cancel()
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
func (c *Crawler) handleResponse(r *http_crawler.ResponseMessage) error {
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

	pageReport.BlockedByRobotstxt = c.robotsChecker.IsBlocked(parsedURL)
	pageReport.InSitemap = c.sitemapStorage.Seen(r.URL)

	if pageReport.Nofollow && !c.options.FollowNofollow {
		return nil
	}

	pageReport.Crawled = true
	c.responseCounter++

	crawlable := [][]*url.URL{
		c.getCrawlableLinks(pageReport),
		c.getResourceURLs(pageReport),
		c.getCrawlableURLs(pageReport),
	}

	urls := []*url.URL{}
	for _, c := range crawlable {
		urls = append(urls, c...)
	}

	for _, t := range urls {
		if c.storage.Seen(t.String()) {
			continue
		}

		c.storage.Add(t.String())

		if !c.options.IgnoreRobotsTxt && c.robotsChecker.IsBlocked(t) {
			c.prStream <- &models.PageReportMessage{
				Crawled:    c.responseCounter,
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

		c.queue.Push(t.String())
	}

	if !pageReport.Noindex || c.options.IncludeNoindex {
		c.prStream <- &models.PageReportMessage{
			PageReport: pageReport,
			HtmlNode:   htmlNode,
			Header:     &r.Response.Header,
			Crawled:    c.responseCounter,
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
			c.queue.Push(v)
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

// Returns a slice with all the crawlable Links from the PageReport's links.
// URLs extracted from internal Links and ExternalLinks are crawlable only if the domain name is allowed and
// if they don't have the "nofollow" attribute. If they have the "nofollow" attribute, they are also considered
// crawlable if the crawler's FollowNofollow option is enabled.
func (c *Crawler) getCrawlableLinks(p *models.PageReport) []*url.URL {
	var urls []*url.URL

	links := append(p.Links, p.ExternalLinks...)
	for _, l := range links {
		if (!l.NoFollow || c.options.FollowNofollow) && c.domainIsAllowed(l.ParsedURL.Host) {
			urls = append(urls, l.ParsedURL)
		}
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
