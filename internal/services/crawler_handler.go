package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"

	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/models"
	"golang.org/x/net/html"
)

type CrawlerHandlerStorage interface {
	SavePageReport(*models.PageReport, int64) (*models.PageReport, error)
}

type CrawlerHandler struct {
	store               CrawlerHandlerStorage
	broker              *Broker
	reportManager       *ReportManager
	client              *crawler.BasicAuthClient
	externalLinksStatus map[string]int
}

type crawlerData struct {
	Depth int
}

func NewCrawlerHandler(s CrawlerHandlerStorage, b *Broker, r *ReportManager, c *crawler.BasicAuthClient) *CrawlerHandler {
	return &CrawlerHandler{
		store:               s,
		broker:              b,
		reportManager:       r,
		client:              c,
		externalLinksStatus: make(map[string]int),
	}
}

func (s *CrawlerHandler) responseCallback(crawl *models.Crawl, p *models.Project, c *crawler.Crawler) crawler.ResponseCallback {
	return func(r *crawler.ResponseMessage) {

		pageReport, htmlNode, err := s.buildPageReport(r)
		if err != nil {
			log.Printf("callback function error: %v", err)
			return
		}

		// Create a requestData object and increase the Depth value
		// according to the data in the responseMessage's Data.
		// If there's no crawlerData in the responseMessage the Depth value
		// is -1 meaning it is undefined.
		requestData := crawlerData{Depth: -1}
		d, ok := r.Data.(crawlerData)
		if ok {
			requestData.Depth = d.Depth + 1
		}

		pageReport.TTFB = r.TTFB
		pageReport.Depth = d.Depth
		pageReport.BlockedByRobotstxt = r.Blocked
		pageReport.InSitemap = r.InSitemap
		pageReport.Crawled = !pageReport.Timeout && (p.FollowNofollow || !pageReport.Nofollow)

		// Add link URLs to the crawler considering the nofollow attribute as well as
		// the projects FollowNoFollow option. In case the URL is blocked by the robots.txt
		// file a new blocked PageReport is saved. Both internal and external links
		// are added as the crawler will discard the domains that are not allowed.
		links := append(pageReport.Links, pageReport.ExternalLinks...)
		for _, l := range links {
			if !l.NoFollow || p.FollowNofollow {
				err := c.AddRequest(&crawler.RequestMessage{URL: l.ParsedURL, Data: requestData})
				if errors.Is(err, crawler.ErrBlockedByRobotstxt) {
					s.saveBlockedPageReport(l.ParsedURL, crawl)
					crawl.BlockedByRobotstxt++
				}
			}
		}

		// Add the indirect URLs such as canonicals, redirects or hreflang URLs to the crawler.
		// In of the URL being blocked by the robots.txt save a new blocked PageReport.
		for _, u := range s.getInderictURLs(pageReport) {
			err := c.AddRequest(&crawler.RequestMessage{URL: u, Data: requestData})
			if errors.Is(err, crawler.ErrBlockedByRobotstxt) {
				s.saveBlockedPageReport(u, crawl)
				crawl.BlockedByRobotstxt++
			}
		}

		// Add the resource URLs to the crawler. If the URL is blocked in the robots.txt
		// Save a new blocked PageReport.
		for _, u := range s.getResourceURLs(pageReport) {
			err := c.AddRequest(&crawler.RequestMessage{URL: u, IgnoreDomain: true, Data: requestData})
			if errors.Is(err, crawler.ErrBlockedByRobotstxt) {
				s.saveBlockedPageReport(u, crawl)
				crawl.BlockedByRobotstxt++
			}
		}

		// Check the external links if the project is set to do so.
		if p.CheckExternalLinks {
			s.checkExternalLinks(pageReport)
		}

		// Save the pageReport if it hasn't the noindex attribute or if the project
		// is set to include the noindexable URLs.
		// If the pageReport is saved correctly create the page issues, otherwise
		// log the error.
		if !pageReport.Noindex || p.IncludeNoindex {
			pageReport, err = s.store.SavePageReport(pageReport, crawl.Id)
			if err == nil {
				s.reportManager.CreatePageIssues(pageReport, htmlNode, &r.Response.Header, crawl)
			} else {
				log.Printf("crawler service: SavePageReport: %v\n", err)
			}
		}

		status := c.GetStatus()
		s.updateStatus(crawl, pageReport)

		s.broker.Publish(fmt.Sprintf("crawl-%d", p.Id), &models.Message{Name: "PageReport", Data: &models.PageReportMessage{
			Crawled:    status.Crawled,
			URL:        r.URL.String(),
			StatusCode: pageReport.StatusCode,
			Crawling:   status.Crawling,
			Discovered: status.Discovered,
		}})
	}
}

// buildPageReport builds a PageReport based on the responseMessage checking for Timeout errors.
func (s *CrawlerHandler) buildPageReport(r *crawler.ResponseMessage) (*models.PageReport, *html.Node, error) {
	// Check if the response caused an error and save a pageReport
	// if it is a Timeout report.
	if r.Error != nil {
		if nErr, ok := r.Error.(net.Error); ok && nErr.Timeout() {
			return &models.PageReport{
				Timeout:   true,
				URL:       r.Response.Request.URL.String(),
				ParsedURL: r.Response.Request.URL,
			}, &html.Node{}, nil
		} else {
			return nil, nil, r.Error
		}
	}

	// Create a new PageReport from the response. If there's a context.DeadlineExceeded
	// error save a pageReport with a timeout.
	pageReport, htmlNode, err := NewFromHTTPResponse(r.Response)
	if err != nil {
		var netErr net.Error
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
			pageReport.Timeout = true
			pageReport.URL = r.Response.Request.URL.String()
			pageReport.ParsedURL = r.Response.Request.URL
		} else {
			return nil, nil, err
		}
	}

	return pageReport, htmlNode, err
}

// saveBlockedPageReport saves a new PageReport with the specified URL and Crawl,
// setting the blockedByRobotstxt field to true.
func (s *CrawlerHandler) saveBlockedPageReport(u *url.URL, crawl *models.Crawl) {
	pageReport := &models.PageReport{
		URL:                u.String(),
		ParsedURL:          u,
		BlockedByRobotstxt: true,
		Crawled:            false,
	}

	_, err := s.store.SavePageReport(pageReport, crawl.Id)
	if err != nil {
		log.Printf("crawler service: SavePageReport: %v\n", err)
	}
}

// updateStatus updates the crawl's data with the total number of crawled URLs, as well as
// the blocked, nofollow, noindex totals as well.
func (s *CrawlerHandler) updateStatus(crawl *models.Crawl, pageReport *models.PageReport) {
	crawl.TotalURLs++

	if pageReport.Noindex {
		crawl.Noindex++
	}

	if pageReport.BlockedByRobotstxt {
		crawl.BlockedByRobotstxt++
	}

	for _, link := range pageReport.Links {
		if link.NoFollow {
			crawl.InternalNoFollowLinks++
		} else {
			crawl.InternalFollowLinks++
		}
	}

	for _, link := range pageReport.ExternalLinks {
		if link.NoFollow {
			crawl.ExternalNoFollowLinks++
		} else {
			crawl.ExternalFollowLinks++
		}
		if link.Sponsored {
			crawl.SponsoredLinks++
		}
		if link.UGC {
			crawl.UGCLinks++
		}
	}
}

// checkExternalLinks makes a HEAD request of the external links in a pageReport
// and checks their status code. It stores the URL's status code in a map to
// avoid requesting the same URL more than once.
func (s *CrawlerHandler) checkExternalLinks(pageReport *models.PageReport) {
	for n, l := range pageReport.ExternalLinks {
		status, ok := s.externalLinksStatus[l.URL]
		if ok {
			pageReport.ExternalLinks[n].StatusCode = status
			continue
		}

		res, err := s.client.Head(l.URL)
		if err != nil {
			log.Printf("external link (%s) error: %v", l.URL, err)
			continue
		}

		s.externalLinksStatus[l.URL] = res.StatusCode
		pageReport.ExternalLinks[n].StatusCode = res.StatusCode
	}
}

// Returns a slice with all the crawlable Links from the PageReport's links.
// URLs extracted from internal Links and ExternalLinks are crawlable only if they don't have
// the "nofollow" attribute. If they have the "nofollow" attribute, they are also considered
// crawlable if the crawler's FollowNofollow option is enabled.
func (s *CrawlerHandler) getInderictURLs(p *models.PageReport) []*url.URL {
	var urls []*url.URL
	var indirect []string

	for _, l := range p.Hreflangs {
		indirect = append(indirect, l.URL)
	}

	indirect = append(indirect, p.Iframes...)

	if p.RedirectURL != "" {
		indirect = append(indirect, p.RedirectURL)
	}

	if p.Canonical != "" {
		indirect = append(indirect, p.Canonical)
	}

	for _, r := range indirect {
		parsed, err := url.Parse(r)
		if err != nil {
			continue
		}

		urls = append(urls, parsed)
	}

	return urls
}

// Returns a slice containing all the resource URLs from a PageReport.
// The resource URLs are always considered crawlable.
func (s *CrawlerHandler) getResourceURLs(p *models.PageReport) []*url.URL {
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
