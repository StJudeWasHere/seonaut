package crawler

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Method int

const (
	// Supported HTTP methods.
	GET Method = iota
	HEAD

	// Random delay in milliseconds.
	// A random delay up to this value is introduced before new HTTP requests.
	randomDelay = 1500

	// Number of threads a queue will use to crawl a project.
	consumerThreads = 2
)

var ErrBlockedByRobotstxt = errors.New("blocked by robots.txt")
var ErrVisited = errors.New("URL already visited")
var ErrDomainNotAllowed = errors.New("domain not allowed")

type Options struct {
	CrawlLimit      int
	IgnoreRobotsTxt bool
	FollowNofollow  bool
	IncludeNoindex  bool
	UserAgent       string
	CrawlSitemap    bool
	AllowSubdomains bool
	AuthUser        string
	AuthPass        string
}

type CrawlerClient interface {
	Get(u string) (*http.Response, error)
	Head(u string) (*http.Response, error)
	Do(req *http.Request) (*http.Response, error)
	GetTTFB(resp *http.Response) int
}

type CrawlerStatus struct {
	Crawled    int
	Crawling   bool
	Discovered int
}

type ResponseCallback func(r *ResponseMessage)

type Crawler struct {
	status           CrawlerStatus
	url              *url.URL
	options          *Options
	queue            *Queue
	cancelQueue      context.CancelFunc
	queueContext     context.Context
	storage          *URLStorage
	sitemapStorage   *URLStorage
	sitemapChecker   *SitemapChecker
	sitemapExists    bool
	sitemapIsBlocked bool
	sitemaps         []string
	robotsChecker    *RobotsChecker
	allowedDomains   map[string]bool
	mainDomain       string
	qStream          chan *RequestMessage
	cancel           context.CancelFunc
	context          context.Context
	client           CrawlerClient
	callback         ResponseCallback
	rStream          chan *ResponseMessage
}

type RequestMessage struct {
	URL          *url.URL
	IgnoreDomain bool
	Method       Method
	Data         interface{}
}

type ResponseMessage struct {
	URL       *url.URL
	Response  *http.Response
	Error     error
	TTFB      int
	Blocked   bool
	InSitemap bool
	Timeout   bool
	Data      interface{}
}

func NewCrawler(url *url.URL, options *Options) *Crawler {
	mainDomain := strings.TrimPrefix(url.Host, "www.")
	httpClient := NewClient(&ClientOptions{
		UserAgent:        options.UserAgent,
		BasicAuthDomains: []string{mainDomain, "www." + mainDomain},
		AuthUser:         options.AuthUser,
		AuthPass:         options.AuthPass,
	})

	storage := NewURLStorage()

	qctx, cancelQueue := context.WithCancel(context.Background())
	q := NewQueue(qctx)

	robotsChecker := NewRobotsChecker(httpClient, options.UserAgent)
	sitemapChecker := NewSitemapChecker(httpClient, options.CrawlLimit)
	qStream := make(chan *RequestMessage)
	ctx, cancel := context.WithCancel(context.Background())

	return &Crawler{
		status:         CrawlerStatus{Crawling: true},
		url:            url,
		options:        options,
		queue:          q,
		cancelQueue:    cancelQueue,
		queueContext:   qctx,
		storage:        storage,
		sitemapStorage: NewURLStorage(),
		sitemapChecker: sitemapChecker,
		robotsChecker:  robotsChecker,
		allowedDomains: map[string]bool{mainDomain: true, "www." + mainDomain: true},
		mainDomain:     mainDomain,
		qStream:        qStream,
		cancel:         cancel,
		context:        ctx,
		client:         httpClient,
		rStream:        make(chan *ResponseMessage),
	}
}

// Polls URLs from the queue and sends them into the qStream channel which is used
// by the httpCrawler to request the URLs.
// The qStream shuts down when the ctx context is done.
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

// OnResponse sets the callback that the crawler will call for every response.
func (c *Crawler) OnResponse(r ResponseCallback) {
	c.callback = r
}

// Crawl starts crawling an URL and sends pagereports of the crawled URLs
// through the pr channel. It will end when there are no more URLs to crawl
// or the MaxPageReports limit is hit.
func (c *Crawler) Start() {
	defer func() {
		c.cancelQueue()
		c.cancel()
	}()

	go c.queueStreamer(c.queueContext)

	c.setupSitemaps()

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

	c.crawl()

	for rm := range c.rStream {
		c.queue.Ack(rm.URL.String())

		rm.InSitemap = c.sitemapStorage.Seen(rm.URL.String())
		rm.Blocked = c.robotsChecker.IsBlocked(rm.URL)
		rm.Timeout = rm.Error != nil

		c.status.Crawled++

		if c.callback != nil {
			c.callback(rm)
		}

		if !c.queue.Active() && c.options.CrawlSitemap && !sitemapLoaded {
			c.queueSitemapURLs()
			sitemapLoaded = true
		}

		if !c.queue.Active() || c.status.Crawled >= c.options.CrawlLimit {
			break
		}
	}
}

// AddRequest processes a request message for the crawler.
// It checks if the URL has already been visited, validates the domain and checks
// if it is blocked in the the robots.txt rules. It returns an error if any of the checks
// fails. Finally, it adds the request to the processing queue.
func (c *Crawler) AddRequest(r *RequestMessage) error {
	if c.storage.Seen(r.URL.String()) {
		return ErrVisited
	}

	c.storage.Add(r.URL.String())

	if !c.domainIsAllowed(r.URL.Host) && !r.IgnoreDomain {
		return ErrDomainNotAllowed
	}

	if !c.options.IgnoreRobotsTxt && c.robotsChecker.IsBlocked(r.URL) {
		return ErrBlockedByRobotstxt
	}

	c.queue.Push(r)

	return nil
}

// GetStatus returns the current cralwer status.
func (c *Crawler) GetStatus() CrawlerStatus {
	c.status.Discovered = c.queue.Count()
	c.status.Crawling = c.context.Err() == nil

	return c.status
}

// Returns true if the sitemap.xml file exists.
func (c *Crawler) SitemapExists() bool {
	return c.sitemapExists
}

// Returns true if the robots.txt file exists.
func (c *Crawler) RobotstxtExists() bool {
	return c.robotsChecker.Exists(c.url)
}

// Returns true if any of the website's sitemaps is blocked in the robots.txt file.
func (c *Crawler) SitemapIsBlocked() bool {
	return c.sitemapIsBlocked
}

// Stops the cralwer by canceling the cralwer context.
func (c *Crawler) Stop() {
	c.cancel()
}

// setupSitemaps checks if any sitemap exists for the crawler's url. It checks the robots file
// as well as the default sitemap location. Afterwards it checks if the sitemap files are blocked
// by the robots file. Any non-blocked sitemap is added to the crawler's sitemaps slice so it can
// be loaded later on.
func (c *Crawler) setupSitemaps() {
	sitemaps := c.robotsChecker.GetSitemaps(c.url)
	nonBlockedSitemaps := []string{}
	if len(sitemaps) == 0 {
		sitemaps = []string{c.url.Scheme + "://" + c.url.Host + "/sitemap.xml"}
	}

	for _, sm := range sitemaps {
		parsedSm, err := url.Parse(sm)
		if err != nil {
			continue
		}

		if c.robotsChecker.IsBlocked(parsedSm) {
			c.sitemapIsBlocked = true
			if !c.options.IgnoreRobotsTxt {
				continue
			}
		}

		nonBlockedSitemaps = append(nonBlockedSitemaps, sm)
	}

	c.sitemaps = nonBlockedSitemaps
	c.sitemapExists = c.sitemapChecker.SitemapExists(sitemaps)
}

// crawl starts the request consumers in goroutines so they can start
// sending requests concurrently.
func (c *Crawler) crawl() {
	go func() {
		defer close(c.rStream)
		wg := new(sync.WaitGroup)
		wg.Add(consumerThreads)

		for i := 0; i < consumerThreads; i++ {
			go func() {
				c.consumer()
				wg.Done()
			}()
		}

		wg.Wait()
	}()
}

// Consumer gets URLs from the qStream until the context is cancelled.
// It adds a random delay between client calls.
func (c *Crawler) consumer() {
	for {
		select {
		case requestMessage := <-c.qStream:
			// Add random delay to avoid overwhelming the servers with requests.
			time.Sleep(time.Duration(rand.Intn(randomDelay)) * time.Millisecond)

			rm := &ResponseMessage{
				URL:  requestMessage.URL,
				Data: requestMessage.Data,
			}

			switch requestMessage.Method {
			case GET:
				rm.Response, rm.Error = c.client.Get(requestMessage.URL.String())
			case HEAD:
				rm.Response, rm.Error = c.client.Head(requestMessage.URL.String())
			}

			// Get the Time To First Byte from the client.
			rm.TTFB = c.client.GetTTFB(rm.Response)

			c.rStream <- rm
		case <-c.context.Done():
			return
		}
	}
}

// Callback to load sitemap URLs into the sitemap storage.
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
			u, err := url.Parse(v)
			if err != nil {
				return
			}

			c.queue.Push(&RequestMessage{URL: u})
		}
	})
}

// Returns true if the crawler is allowed to crawl the domain, checking the allowedDomains slice.
// If the AllowSubdomains option is set, returns true the given domain is a subdomain of the
// crawlers's base domain.
func (c *Crawler) domainIsAllowed(d string) bool {
	_, ok := c.allowedDomains[d]
	if ok {
		return true
	}

	if c.options.AllowSubdomains && strings.HasSuffix(d, c.mainDomain) {
		return true
	}

	return false
}
