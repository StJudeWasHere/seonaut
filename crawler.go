package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
)

const (
	consumerThreads = 2
	storageMaxSize  = 10000
	MaxPageReports  = 10000
)

type Crawler struct{}

func startCrawler(p Project) int {
	var crawled int

	pageReport := make(chan PageReport)
	c := &Crawler{}

	u, err := url.Parse(p.URL)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	cid := saveCrawl(p)

	go c.Crawl(u, p.IgnoreRobotsTxt, pageReport)

	for r := range pageReport {
		crawled++
		savePageReport(&r, cid)
	}

	saveEndCrawl(cid, time.Now())
	fmt.Printf("%d pages crawled.\n", crawled)

	return int(cid)
}

func (c *Crawler) Crawl(u *url.URL, ignoreRobotsTxt bool, pr chan<- PageReport) {
	defer close(pr)

	q, _ := queue.New(
		consumerThreads,
		&queue.InMemoryQueueStorage{MaxSize: storageMaxSize},
	)

	var responseCounter int

	handleResponse := func(r *colly.Response) {
		if responseCounter >= MaxPageReports {
			return
		}

		pageReport := NewPageReport(r.Request.URL, r.StatusCode, r.Headers, r.Body)
		pr <- *pageReport

		responseCounter++

		for _, l := range pageReport.Links {
			if strings.Contains(l.Rel, "nofollow") {
				continue
			}

			q.AddURL(r.Request.AbsoluteURL(l.URL))
		}

		if pageReport.RedirectURL != "" {
			q.AddURL(r.Request.AbsoluteURL(pageReport.RedirectURL))
		}

		for _, l := range pageReport.Scripts {
			q.AddURL(r.Request.AbsoluteURL(l))
		}

		for _, l := range pageReport.Styles {
			q.AddURL(r.Request.AbsoluteURL(l))
		}

		for _, l := range pageReport.Images {
			q.AddURL(r.Request.AbsoluteURL(l.URL))
		}

		for _, l := range pageReport.Hreflangs {
			q.AddURL(r.Request.AbsoluteURL(l.URL))
		}

		if pageReport.Canonical != "" {
			q.AddURL(r.Request.AbsoluteURL(pageReport.Canonical))
		}
	}

	co := colly.NewCollector(
		colly.AllowedDomains(u.Host),
		colly.UserAgent(config.CrawlerAgent),
		func(c *colly.Collector) {
			c.IgnoreRobotsTxt = ignoreRobotsTxt
		},
	)

	co.OnRequest(func(r *colly.Request) {
		// fmt.Printf("Visiting %s\n", r.URL.String())
	})

	co.OnResponse(handleResponse)

	co.OnError(func(r *colly.Response, err error) {
		if r.StatusCode > 0 && r.Headers != nil {
			handleResponse(r)
		}
	})

	co.SetRedirectHandler(func(r *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	})

	if u.Path == "" {
		u.Path = "/"
	}

	q.AddURL(u.String())

	q.Run(co)
}
