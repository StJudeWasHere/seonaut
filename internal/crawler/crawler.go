package crawler

import (
	"net/http"
	"net/url"
	"sync"

	"github.com/gocolly/colly/v2"
	"github.com/temoto/robotstxt"
)

const (
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

	robotsMap       map[string]*robotstxt.RobotsData
	rlock           *sync.RWMutex
	storage         *URLStorage
	sitemapChecker  *SitemapChecker
	sitemapExists   bool
	robotstxtExists bool
	plock           *sync.RWMutex
	sitemapsMap     []string
	responseCounter int

	que *que
	pr  chan<- PageReport
}

func NewCrawler(url *url.URL, agent string, max int, irobots, fnofollow, inoindex, crawlSitemap bool) *Crawler {
	return &Crawler{
		URL:             url,
		MaxPageReports:  max,
		IgnoreRobotsTxt: irobots,
		FollowNofollow:  fnofollow,
		IncludeNoindex:  inoindex,
		UserAgent:       agent,
		CrawlSitemap:    crawlSitemap,

		sitemapsMap:    []string{url.Scheme + "://" + url.Host + "/sitemap.xml"},
		robotsMap:      make(map[string]*robotstxt.RobotsData),
		rlock:          &sync.RWMutex{},
		storage:        NewURLStorage(),
		sitemapChecker: NewSitemapChecker(),
		plock:          &sync.RWMutex{},

		que: NewQueue(),
	}
}

// Gets an URL and handles the response with the responseHandler method
func (c *Crawler) get(u string) {
	co := colly.NewCollector(
		colly.Async(false),
	)
	co.UserAgent = c.UserAgent
	co.IgnoreRobotsTxt = c.IgnoreRobotsTxt

	// Redirect handler
	handleRedirect := func(r *http.Request, via []*http.Request) error {
		for _, v := range via {
			if v.URL.Path == "/robots.txt" {
				return nil
			}
		}

		return http.ErrUseLastResponse
	}

	co.OnResponse(c.responseHandler)
	co.SetRedirectHandler(handleRedirect)
	co.OnError(func(r *colly.Response, err error) {
		if r.StatusCode > 0 && r.Headers != nil {
			c.responseHandler(r)
			return
		}
	})

	co.Visit(u)
	co.Wait()
}

// Crawl starts crawling an URL and sends pagereports of the crawled URLs
// through the pr channel. It will end when there are no more URLs to crawl
// or the MaxPageReports limit is hit.
func (c *Crawler) Crawl(pr chan<- PageReport) {
	defer close(pr)

	c.pr = pr

	robot, err := c.getRobotsMap(c.URL)
	if err == nil && robot != nil {
		c.sitemapsMap = c.removeDuplicates(append(c.sitemapsMap, robot.Sitemaps...))
	}

	c.sitemapExists = c.sitemapChecker.SitemapExists(c.sitemapsMap)

	if c.URL.Path == "" {
		c.URL.Path = "/"
	}

	if c.CrawlSitemap && c.sitemapExists {
		go func() {
			c.sitemapChecker.ParseSitemaps(c.sitemapsMap, func(u string) {
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
			})
		}()
	}

	c.que.Push(c.URL.String())
	c.storage.Add(c.URL.String())

	wg := new(sync.WaitGroup)
	wg.Add(consumerThreads)

	for i := 0; i < consumerThreads; i++ {
		go func(w *sync.WaitGroup, n int) {
			for {
				url, ok := c.que.Poll()
				if !ok {
					break
				}
				c.get(url.(string))
			}
			w.Done()
		}(wg, i)
	}

	wg.Wait()
}

// Check if URL is blocked by robots.txt
func (c *Crawler) isBlockedByRobotstxt(u *url.URL) bool {
	robot, err := c.getRobotsMap(u)
	if err != nil || robot == nil {
		return true
	}

	path := u.EscapedPath()
	if u.RawQuery != "" {
		path += "?" + u.Query().Encode()
	}

	return !robot.TestAgent(path, c.UserAgent)
}

// Returns a RobotsData checking if it has already been created and stored in the robotsMap
func (c *Crawler) getRobotsMap(u *url.URL) (*robotstxt.RobotsData, error) {
	c.rlock.RLock()
	robot, ok := c.robotsMap[u.Host]
	c.rlock.RUnlock()

	if !ok {
		resp, err := http.Get(u.Scheme + "://" + u.Host + "/robots.txt")
		if err != nil {
			c.rlock.Lock()
			c.robotsMap[u.Host] = nil
			c.rlock.Unlock()

			return nil, err
		}
		defer resp.Body.Close()

		if u.Host == c.URL.Host && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			c.robotstxtExists = true
		}

		robot, err = robotstxt.FromResponse(resp)
		if err != nil {
			c.rlock.Lock()
			c.robotsMap[u.Host] = nil
			c.rlock.Unlock()

			return nil, err
		}

		c.rlock.Lock()
		c.robotsMap[u.Host] = robot
		c.rlock.Unlock()
	}

	return robot, nil
}

// Returns true if the sitemap.xml file exists
func (c *Crawler) SitemapExists() bool {
	return c.sitemapExists
}

// Returns true if the robots.txt file exists
func (c *Crawler) RobotstxtExists() bool {
	return c.robotstxtExists
}

// Remove duplicate strings from slice
func (c *Crawler) removeDuplicates(m []string) []string {
	s := make(map[string]bool)
	var unique []string

	for _, str := range m {
		if _, ok := s[str]; !ok {
			s[str] = true
			unique = append(unique, str)
		}
	}

	return unique
}

// Returns true if the page report limit has been hit
func (c *Crawler) isLimitHit() bool {
	c.plock.RLock()
	defer c.plock.RUnlock()

	return c.responseCounter >= c.MaxPageReports
}

// Handles the HTTP response
func (c *Crawler) responseHandler(r *colly.Response) {
	defer c.que.Ack(r.Request.AbsoluteURL(r.Request.URL.String()))

	if c.isLimitHit() {
		return
	}

	u := r.Request.URL
	pageReport := NewPageReport(u, r.StatusCode, r.Headers, r.Body)
	pageReport.BlockedByRobotstxt = c.isBlockedByRobotstxt(u)

	if pageReport.Noindex == false || c.IncludeNoindex == true {
		pageReport.Crawled = true
		c.plock.Lock()
		c.responseCounter++
		c.plock.Unlock()
	}

	c.pr <- *pageReport

	if pageReport.Nofollow == true && c.FollowNofollow == false {
		return
	}

	var toVisit []*url.URL

	for _, l := range pageReport.Links {
		if l.NoFollow && c.FollowNofollow == false {
			continue
		}

		toVisit = append(toVisit, l.ParsedURL)
	}

	if pageReport.RedirectURL != "" {
		parsed, err := url.Parse(pageReport.RedirectURL)
		if err == nil {
			toVisit = append(toVisit, parsed)
		}
	}

	for _, l := range pageReport.Hreflangs {
		parsed, err := url.Parse(l.URL)
		if err != nil {
			continue
		}
		toVisit = append(toVisit, parsed)
	}

	if pageReport.Canonical != "" {
		parsed, err := url.Parse(pageReport.Canonical)
		if err == nil {
			toVisit = append(toVisit, parsed)
		}
	}

	var resources []string

	for _, l := range pageReport.Scripts {
		resources = append(resources, r.Request.AbsoluteURL(l))
	}

	for _, l := range pageReport.Styles {
		resources = append(resources, r.Request.AbsoluteURL(l))
	}

	for _, l := range pageReport.Images {
		resources = append(resources, r.Request.AbsoluteURL(l.URL))
	}

	for _, v := range resources {
		t, err := url.Parse(v)
		if err != nil {
			continue
		}
		toVisit = append(toVisit, t)
	}

	for _, t := range toVisit {
		if c.storage.Seen(r.Request.AbsoluteURL(t.String())) {
			continue
		}

		if c.IgnoreRobotsTxt == false && c.isBlockedByRobotstxt(t) {
			p := &PageReport{
				URL:                t.String(),
				ParsedURL:          t,
				Crawled:            false,
				BlockedByRobotstxt: true,
			}

			c.pr <- *p
		}

		if c.IgnoreRobotsTxt == true || c.isBlockedByRobotstxt(t) == false {
			c.que.Push(r.Request.AbsoluteURL(t.String()))
		}
		c.storage.Add(r.Request.AbsoluteURL(t.String()))
	}
}
