package crawler

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/temoto/robotstxt"
)

const (
	// Number of threads a queue will use to crawl a project
	consumerThreads = 2

	// Max capacity of a queue
	storageMaxSize = 10000
)

type Crawler struct {
	URL             *url.URL
	MaxPageReports  int
	IgnoreRobotsTxt bool
	FollowNofollow  bool
	IncludeNoindex  bool
	UserAgent       string
	CrawlSitemap    bool

	robotsMap map[string]*robotstxt.RobotsData
	rlock     *sync.RWMutex

	storage        *URLStorage
	sitemapChecker *SitemapChecker

	sitemapExists   bool
	robotstxtExists bool

	sitemapsMap []string
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

		robotsMap: make(map[string]*robotstxt.RobotsData),
		rlock:     &sync.RWMutex{},

		storage:        NewURLStorage(),
		sitemapChecker: NewSitemapChecker(),
	}
}

// Crawl starts crawling an URL and sends pagereports of the crawled URLs
// through the pr channel. It will end when there are no more URLs to crawl
// or the MaxPageReports limit is hit.
func (c *Crawler) Crawl(pr chan<- PageReport) {
	defer close(pr)

	robot := c.getRobotsMap(c.URL)
	c.sitemapsMap = append(robot.Sitemaps, c.URL.Scheme+"://"+c.URL.Host+"/sitemap.xml")
	c.sitemapExists = c.sitemapChecker.SitemapExists(c.sitemapsMap)

	q, _ := queue.New(
		consumerThreads,
		&queue.InMemoryQueueStorage{MaxSize: storageMaxSize},
	)

	var responseCounter int

	// Crawl the www and non-www domain
	allowedDomains := []string{c.URL.Host}
	if strings.HasPrefix(c.URL.Host, "www.") {
		allowedDomains = append(allowedDomains, c.URL.Host[4:])
	} else {
		allowedDomains = append(allowedDomains, "www."+c.URL.Host)
	}

	// Links collector
	co := colly.NewCollector()
	co.UserAgent = c.UserAgent
	co.AllowedDomains = allowedDomains
	co.IgnoreRobotsTxt = c.IgnoreRobotsTxt

	// Resources collector allows any domain
	cor := colly.NewCollector()
	cor.UserAgent = c.UserAgent
	cor.IgnoreRobotsTxt = c.IgnoreRobotsTxt

	// Resources response hlandler
	handleResourceResponse := func(r *colly.Response) {
		if responseCounter >= c.MaxPageReports {
			return
		}
		url := r.Request.URL
		pageReport := NewPageReport(url, r.StatusCode, r.Headers, r.Body)
		pageReport.BlockedByRobotstxt = c.isBlockedByRobotstxt(url)
		pageReport.Crawled = true

		pr <- *pageReport
		responseCounter++
	}

	// Links response handler
	handleResponse := func(r *colly.Response) {
		if responseCounter >= c.MaxPageReports {
			return
		}

		u := r.Request.URL
		pageReport := NewPageReport(u, r.StatusCode, r.Headers, r.Body)
		pageReport.BlockedByRobotstxt = c.isBlockedByRobotstxt(u)

		if pageReport.Noindex == false || c.IncludeNoindex == true {
			pageReport.Crawled = true
			responseCounter++
		}

		pr <- *pageReport

		if strings.Contains(pageReport.Robots, "nofollow") && c.FollowNofollow == false {
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

		for _, t := range toVisit {
			if c.IgnoreRobotsTxt == false && c.isBlockedByRobotstxt(t) && !c.storage.Seen(t.String()) {
				c.storage.Add(t.String())

				p := &PageReport{
					URL:                t.String(),
					ParsedURL:          t,
					Crawled:            false,
					BlockedByRobotstxt: true,
				}

				pr <- *p
			}

			q.AddURL(r.Request.AbsoluteURL(t.String()))
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

		if len(resources) > 0 {
			qr, _ := queue.New(
				consumerThreads,
				&queue.InMemoryQueueStorage{MaxSize: storageMaxSize},
			)

			for _, v := range resources {
				visited, err := co.HasVisited(v)
				if err != nil || visited == true {
					continue
				}

				t, err := url.Parse(v)
				if err != nil {
					continue
				}

				if c.IgnoreRobotsTxt == false && c.isBlockedByRobotstxt(t) && !c.storage.Seen(t.String()) {
					c.storage.Add(t.String())

					p := &PageReport{
						URL:                t.String(),
						ParsedURL:          t,
						Crawled:            false,
						BlockedByRobotstxt: true,
					}

					pr <- *p
				}

				qr.AddURL(v)
			}

			qr.Run(cor)
		}
	}

	// Redirect handler
	handleRedirect := func(r *http.Request, via []*http.Request) error {
		for _, v := range via {
			if v.URL.Path == "/robots.txt" {
				return nil
			}
		}

		return http.ErrUseLastResponse
	}

	co.OnResponse(handleResponse)
	co.SetRedirectHandler(handleRedirect)
	co.OnError(func(r *colly.Response, err error) {
		if r.StatusCode > 0 && r.Headers != nil {
			handleResponse(r)
		}
	})

	cor.OnResponse(handleResourceResponse)
	cor.SetRedirectHandler(handleRedirect)
	cor.OnError(func(r *colly.Response, err error) {
		if r.StatusCode > 0 && r.Headers != nil {
			handleResourceResponse(r)
		}
	})

	if c.URL.Path == "" {
		c.URL.Path = "/"
	}

	if c.CrawlSitemap && c.sitemapExists {
		go func() {
			c.sitemapChecker.ParseSitemaps(c.sitemapsMap, func(u string) {
				if responseCounter >= c.MaxPageReports {
					return
				}

				l, err := url.Parse(u)
				if err != nil {
					return
				}

				if l.Path == "/" {
					l.Path = "/"
				}

				q.AddURL(l.String())
			})
		}()
	}

	q.AddURL(c.URL.String())
	q.Run(co)
}

// Check if URL is blocked by robots.txt
func (c *Crawler) isBlockedByRobotstxt(u *url.URL) bool {
	robot := c.getRobotsMap(u)
	if robot == nil {
		return true
	}

	path := u.EscapedPath()
	if u.RawQuery != "" {
		path += "?" + u.Query().Encode()
	}

	return !robot.TestAgent(path, c.UserAgent)
}

// Returns a RobotsData checking if it has already been created and stored in the robotsMap
func (c *Crawler) getRobotsMap(u *url.URL) *robotstxt.RobotsData {
	c.rlock.RLock()
	robot, ok := c.robotsMap[u.Host]
	c.rlock.RUnlock()

	if !ok {
		resp, err := http.Get(u.Scheme + "://" + u.Host + "/robots.txt")
		if err != nil {
			c.rlock.Lock()
			c.robotsMap[u.Host] = nil
			c.rlock.Unlock()

			return nil
		}
		defer resp.Body.Close()

		if u.Host == c.URL.Host && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			c.robotstxtExists = true
		}

		robot, err = robotstxt.FromResponse(resp)
		if err != nil {
			log.Printf("getRobotsMap: %v\n", err)
		}

		c.rlock.Lock()
		c.robotsMap[u.Host] = robot
		c.rlock.Unlock()
	}

	return robot
}

// Returns true if the sitemap.xml file exists
func (c *Crawler) SitemapExists() bool {
	return c.sitemapExists
}

// Returns true if the robots.txt file exists
func (c *Crawler) RobotstxtExists() bool {
	return c.robotstxtExists
}
