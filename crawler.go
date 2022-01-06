package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
)

const (
	userAgent       = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Safari/605.1.15"
	consumerThreads = 2
	storageMaxSize  = 10000
)

type Crawler struct{}

func (c *Crawler) Crawl(u *url.URL, pr chan<- PageReport) {
	defer close(pr)

	q, _ := queue.New(
		consumerThreads,
		&queue.InMemoryQueueStorage{MaxSize: storageMaxSize},
	)

	handleResponse := func(r *colly.Response) {
		pageReport := NewPageReport(r.Request.URL, r.StatusCode, r.Headers, r.Body)
		pr <- *pageReport

		for _, l := range pageReport.Links {
			if l.External {
				continue
			}
			u := l.URL
			if strings.Contains(u, "#") {
				u = strings.Split(u, "#")[0]
			}

			q.AddURL(r.Request.AbsoluteURL(u))
		}

		if pageReport.RedirectURL != "" {
			q.AddURL(pageReport.RedirectURL)
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

		if pageReport.Refresh != "" {
			u := strings.Split(pageReport.Refresh, ";")
			if len(u) > 1 && strings.ToLower(u[1][:4]) == "url=" {
				url := strings.ReplaceAll(u[1][4:], "'", "")
				q.AddURL(r.Request.AbsoluteURL(url))
			}
		}
	}

	co := colly.NewCollector(colly.AllowedDomains(u.Host), colly.UserAgent(userAgent))

	co.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting: ", r.URL.String())
	})

	co.OnResponse(handleResponse)

	co.OnError(func(r *colly.Response, err error) {
		handleResponse(r)
	})

	co.SetRedirectHandler(func(r *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	})

	q.AddURL(u.String())

	q.Run(co)
}
