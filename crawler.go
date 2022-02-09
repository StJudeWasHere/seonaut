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
	RendertronURL   = "http://127.0.0.1:3000/render/"
)

type Crawler struct{}

func startCrawler(p Project, agent string, datastore *datastore) int {
	var crawled int

	pageReport := make(chan PageReport)
	c := &Crawler{}

	u, err := url.Parse(p.URL)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	cid := datastore.saveCrawl(p)

	go c.Crawl(u, p.IgnoreRobotsTxt, p.UseJS, agent, pageReport)

	for r := range pageReport {
		crawled++
		datastore.savePageReport(&r, cid)
	}

	datastore.saveEndCrawl(cid, time.Now())
	fmt.Printf("%d pages crawled.\n", crawled)

	return int(cid)
}

func (c *Crawler) Crawl(u *url.URL, ignoreRobotsTxt, useJS bool, agent string, pr chan<- PageReport) {
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

		us := r.Request.URL.String()
		url := r.Request.URL
		if useJS == true {
			us = us[len(RendertronURL):]
			url, _ = url.Parse(us)
		}
		pageReport := NewPageReport(url, r.StatusCode, r.Headers, r.Body)
		pr <- *pageReport

		responseCounter++

		for _, l := range pageReport.Links {
			if strings.Contains(l.Rel, "nofollow") {
				continue
			}

			lurl := r.Request.AbsoluteURL(l.URL)
			if useJS == true {
				lurl = RendertronURL + lurl
			}

			q.AddURL(lurl)
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

	var nonWWWHost string
	var WWWHost string
	if strings.HasPrefix(u.Host, "www.") {
		WWWHost = u.Host
		nonWWWHost = u.Host[4:]
	} else {
		WWWHost = "www." + u.Host
		nonWWWHost = u.Host
	}

	co := colly.NewCollector(
		colly.AllowedDomains(WWWHost, nonWWWHost, "127.0.0.1"),
		colly.UserAgent(agent),
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
		for _, v := range via {
			if v.URL.Path == "/robots.txt" {
				return nil
			}
		}

		return http.ErrUseLastResponse
	})

	if u.Path == "" {
		u.Path = "/"
	}

	us := u.String()
	if useJS == true {
		us = RendertronURL + us
		n, _ := time.ParseDuration("30s")
		co.SetRequestTimeout(n)
	}

	q.AddURL(us)

	q.Run(co)
}
