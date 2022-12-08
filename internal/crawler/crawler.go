package crawler

import (
	"context"
	"net/url"
	"strings"

	"github.com/stjudewashere/seonaut/internal/http_crawler"
	"github.com/stjudewashere/seonaut/internal/pagereport"
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
	sitemapChecker  *SitemapChecker
	sitemapExists   bool
	sitemaps        []string
	robotstxtExists bool
	responseCounter int
	robotsChecker   *RobotsChecker
	prStream        chan *PageReportMessage
	allowedDomains  map[string]bool
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

	robotsChecker := NewRobotsChecker(options.UserAgent)

	sitemaps := robotsChecker.GetSitemaps(url)
	if len(sitemaps) == 0 {
		sitemaps = []string{url.Scheme + "://" + url.Host + "/sitemap.xml"}
	}

	sitemapChecker := NewSitemapChecker(options.MaxPageReports)
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
		prStream:        make(chan *PageReportMessage),
		qStream:         qStream,
		httpCrawler: http_crawler.New(
			http_crawler.NewClient(&http_crawler.ClientOptions{
				UserAgent: options.UserAgent,
				BasicAuth: options.BasicAuth,
				AuthUser:  options.AuthUser,
				AuthPass:  options.AuthPass,
			}),
			qStream,
		),
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
func (c *Crawler) Stream() <-chan *PageReportMessage {
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
			continue
		}

		if c.queue.Active() == false && c.options.CrawlSitemap && sitemapLoaded == false {
			c.queueSitemapURLs()
			sitemapLoaded = true
		}

		if c.queue.Active() == false || c.responseCounter >= c.options.MaxPageReports {
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

	pageReport, err := pagereport.NewPageReportFromHTTPResponse(r.Response)
	if err != nil {
		return err
	}

	parsedURL, err := url.Parse(r.URL)
	if err != nil {
		return err
	}

	pageReport.BlockedByRobotstxt = c.robotsChecker.IsBlocked(parsedURL)
	pageReport.InSitemap = c.sitemapStorage.Seen(r.URL)

	if pageReport.Noindex == false || c.options.IncludeNoindex == true {
		pageReport.Crawled = true
		c.responseCounter++
	}

	if pageReport.Nofollow == true && c.options.FollowNofollow == false {
		return nil
	}

	for _, t := range c.getCrawlableURLs(pageReport) {
		if c.storage.Seen(t.String()) == true {
			continue
		}

		c.storage.Add(t.String())

		if c.options.IgnoreRobotsTxt == false && c.robotsChecker.IsBlocked(t) {
			c.prStream <- &PageReportMessage{
				Crawled:    c.responseCounter,
				Discovered: c.queue.Count(),
				PageReport: &pagereport.PageReport{
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

	message := &PageReportMessage{
		PageReport: pageReport,
		Crawled:    c.responseCounter,
		Discovered: c.queue.Count(),
	}

	c.prStream <- message

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

	if c.options.AllowSubdomains && strings.HasSuffix(s, c.url.Host) {
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
		if c.storage.Seen(v) == false {
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

// Returns all the crawlable URLs found in the HTML document.
// URLs extracted from the PageReport's Scripts, Styles, Images, Audios and Videos are always considered crawlable.
// HrefLangs, Iframes, RedirectURLs and Canonical URLs are crawlable only if their the domain name is allowed.
// URLs extracted from internal Links and ExternalLinks are crawlable only if the domain name is allowed and
// if they don't have the "nofollow" attribute. If they have the "nofollow" attribute, they are also considered
// crawlable if the crawler's FollowNofollow option is enabled.
func (c *Crawler) getCrawlableURLs(p *pagereport.PageReport) []*url.URL {
	var urls []*url.URL
	var resources []string

	for _, l := range p.Links {
		if (!l.NoFollow || c.options.FollowNofollow) && c.domainIsAllowed(l.ParsedURL.Host) {
			urls = append(urls, l.ParsedURL)
		}
	}

	for _, l := range p.ExternalLinks {
		if (!l.NoFollow || c.options.FollowNofollow) && c.domainIsAllowed(l.ParsedURL.Host) {
			urls = append(urls, l.ParsedURL)
		}
	}

	for _, l := range p.Hreflangs {
		parsed, err := url.Parse(l.URL)
		if err != nil {
			continue
		}

		if c.domainIsAllowed(parsed.Host) {
			urls = append(urls, parsed)
		}
	}

	for _, l := range p.Iframes {
		parsed, err := url.Parse(l)
		if err != nil {
			continue
		}

		if c.domainIsAllowed(parsed.Host) {
			urls = append(urls, parsed)
		}
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
