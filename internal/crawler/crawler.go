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

	robotsMap map[string]*robotstxt.RobotsData
	lock      *sync.RWMutex
}

func NewCrawler(url *url.URL, agent string, max int, irobots, fnofollow, inoindex bool) *Crawler {
	return &Crawler{
		URL:             url,
		MaxPageReports:  max,
		IgnoreRobotsTxt: irobots,
		FollowNofollow:  fnofollow,
		IncludeNoindex:  inoindex,
		UserAgent:       agent,

		robotsMap: make(map[string]*robotstxt.RobotsData),
		lock:      &sync.RWMutex{},
	}
}

// Crawl starts crawling an URL and sends pagereports of the crawled URLs
// through the pr channel. It will end when there are no more URLs to crawl
// or the MaxPageReports limit is hit.
func (c *Crawler) Crawl(pr chan<- PageReport) {
	defer close(pr)

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

		pr <- *pageReport
		responseCounter++
	}

	// Links response handler
	handleResponse := func(r *colly.Response) {
		if responseCounter >= c.MaxPageReports {
			return
		}

		url := r.Request.URL
		pageReport := NewPageReport(url, r.StatusCode, r.Headers, r.Body)
		pageReport.BlockedByRobotstxt = c.isBlockedByRobotstxt(url)

		if pageReport.Noindex == false || c.IncludeNoindex == true {
			pr <- *pageReport
			responseCounter++
		}

		if strings.Contains(pageReport.Robots, "nofollow") && c.FollowNofollow == false {
			return
		}

		for _, l := range pageReport.Links {
			if l.NoFollow && c.FollowNofollow == false {
				continue
			}

			q.AddURL(r.Request.AbsoluteURL(l.URL))
		}

		if pageReport.RedirectURL != "" {
			q.AddURL(r.Request.AbsoluteURL(pageReport.RedirectURL))
		}

		for _, l := range pageReport.Hreflangs {
			q.AddURL(r.Request.AbsoluteURL(l.URL))
		}

		if pageReport.Canonical != "" {
			q.AddURL(r.Request.AbsoluteURL(pageReport.Canonical))
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
				if err != nil {
					log.Printf("crawler: collector has visited: %v\n", err)
					continue
				}

				if visited == false {
					qr.AddURL(v)
				}
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

	us := c.URL.String()

	q.AddURL(us)
	q.Run(co)
}

func (c *Crawler) isBlockedByRobotstxt(u *url.URL) bool {
	c.lock.RLock()
	robot, ok := c.robotsMap[u.Host]
	c.lock.RUnlock()

	if !ok {
		resp, err := http.Get(u.Scheme + "://" + u.Host + "/robots.txt")
		if err != nil {
			c.lock.Lock()
			c.robotsMap[u.Host] = nil
			c.lock.Unlock()

			return true
		}
		defer resp.Body.Close()

		robot, err = robotstxt.FromResponse(resp)
		if err != nil {
			log.Println(err)
		}

		c.lock.Lock()
		c.robotsMap[u.Host] = robot
		c.lock.Unlock()
	}

	if robot == nil {
		return true
	}

	path := u.EscapedPath()
	if u.RawQuery != "" {
		path += "?" + u.Query().Encode()
	}

	return !robot.TestAgent(path, c.UserAgent)
}
