package main

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/microcosm-cc/bluemonday"
)

const (
	consumerThreads        = 2
	storageMaxSize         = 10000
	MaxPageReports         = 300
	AdvancedMaxPageReports = 5000
	RendertronURL          = "http://127.0.0.1:3000/render/"
)

type Crawler struct {
	URL             *url.URL
	MaxPageReports  int
	UseJS           bool
	IgnoreRobotsTxt bool
	UserAgent       string
	sanitizer       *bluemonday.Policy
}

func startCrawler(p Project, agent string, advanced bool, datastore *datastore, sanitizer *bluemonday.Policy) int {
	var totalURLs int
	var max int

	if advanced {
		max = AdvancedMaxPageReports
	} else {
		max = MaxPageReports
	}

	u, err := url.Parse(p.URL)
	if err != nil {
		log.Printf("startCrawler: %s %v\n", p.URL, err)
		return 0
	}

	c := &Crawler{
		URL:             u,
		MaxPageReports:  max,
		UseJS:           p.UseJS,
		IgnoreRobotsTxt: p.IgnoreRobotsTxt,
		UserAgent:       agent,
		sanitizer:       sanitizer,
	}

	cid := datastore.saveCrawl(p)

	pageReport := make(chan PageReport)
	go c.Crawl(pageReport)

	for r := range pageReport {
		totalURLs++
		datastore.savePageReport(&r, cid)
	}

	datastore.saveEndCrawl(cid, time.Now(), totalURLs)
	log.Printf("%d pages crawled.\n", totalURLs)

	return int(cid)
}

func (c *Crawler) Crawl(pr chan<- PageReport) {
	defer close(pr)

	q, _ := queue.New(
		consumerThreads,
		&queue.InMemoryQueueStorage{MaxSize: storageMaxSize},
	)

	var responseCounter int
	cor := colly.NewCollector(
		colly.UserAgent(c.UserAgent),
		func(co *colly.Collector) {
			co.IgnoreRobotsTxt = c.IgnoreRobotsTxt
		},
	)

	handleResourceResponse := func(r *colly.Response) {
		if responseCounter >= c.MaxPageReports {
			return
		}
		url := r.Request.URL
		pageReport := NewPageReport(url, r.StatusCode, r.Headers, r.Body, c.sanitizer)
		pr <- *pageReport
		responseCounter++
	}

	handleResponse := func(r *colly.Response) {
		if responseCounter >= c.MaxPageReports {
			return
		}

		us := r.Request.URL.String()
		url := r.Request.URL
		if c.UseJS == true {
			us = us[len(RendertronURL):]
			url, _ = url.Parse(us)
		}
		pageReport := NewPageReport(url, r.StatusCode, r.Headers, r.Body, c.sanitizer)

		if strings.Contains(pageReport.Robots, "noindex") {
			return
		}

		pr <- *pageReport
		responseCounter++

		if strings.Contains(pageReport.Robots, "nofollow") {
			return
		}

		for _, l := range pageReport.Links {
			if l.NoFollow {
				continue
			}

			lurl := r.Request.AbsoluteURL(l.URL)
			if c.UseJS == true {
				lurl = RendertronURL + lurl
			}

			q.AddURL(lurl)
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
				qr.AddURL(v)
			}

			qr.Run(cor)
		}
	}

	var nonWWWHost string
	var WWWHost string
	if strings.HasPrefix(c.URL.Host, "www.") {
		WWWHost = c.URL.Host
		nonWWWHost = c.URL.Host[4:]
	} else {
		WWWHost = "www." + c.URL.Host
		nonWWWHost = c.URL.Host
	}

	co := colly.NewCollector(
		colly.AllowedDomains(WWWHost, nonWWWHost, "127.0.0.1"),
		colly.UserAgent(c.UserAgent),
		func(co *colly.Collector) {
			co.IgnoreRobotsTxt = c.IgnoreRobotsTxt
		},
	)

	co.OnResponse(handleResponse)

	co.OnError(func(r *colly.Response, err error) {
		if r.StatusCode > 0 && r.Headers != nil {
			handleResponse(r)
		}
	})

	co.SetRedirectHandler(func(r *http.Request, via []*http.Request) error {
		for _, v := range via {
			if v.URL.Path == "/robots.txt" {
				return nil
			}
		}

		return http.ErrUseLastResponse
	})

	cor.OnResponse(handleResourceResponse)

	cor.OnError(func(r *colly.Response, err error) {
		if r.StatusCode > 0 && r.Headers != nil {
			handleResourceResponse(r)
		}
	})

	cor.SetRedirectHandler(func(r *http.Request, via []*http.Request) error {
		for _, v := range via {
			if v.URL.Path == "/robots.txt" {
				return nil
			}
		}

		return http.ErrUseLastResponse
	})

	if c.URL.Path == "" {
		c.URL.Path = "/"
	}

	us := c.URL.String()
	if c.UseJS == true {
		us = RendertronURL + us
		n, _ := time.ParseDuration("30s")
		co.SetRequestTimeout(n)
	}

	q.AddURL(us)
	q.Run(co)
}
