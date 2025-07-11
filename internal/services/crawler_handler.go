package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/urlutils"
	"golang.org/x/net/html"
)

type CrawlerHandlerRepository interface {
	SavePageReport(*models.PageReport, int64) (*models.PageReport, error)
}

type CrawlerHandler struct {
	repository          CrawlerHandlerRepository
	broker              *Broker
	reportManager       *ReportManager
	externalLinksStatus map[string]int
}

type crawlerData struct {
	Depth int
}

type Archiver interface {
	AddRecord(*http.Response)
}

func NewCrawlerHandler(r CrawlerHandlerRepository, b *Broker, m *ReportManager) *CrawlerHandler {
	return &CrawlerHandler{
		repository:          r,
		broker:              b,
		reportManager:       m,
		externalLinksStatus: make(map[string]int),
	}
}

func (s *CrawlerHandler) archiveWrapper(callback crawler.ResponseCallback, a Archiver) crawler.ResponseCallback {
	return func(r *crawler.ResponseMessage) {
		if r.Error == nil && a != nil {
			a.AddRecord(r.Response)
		}
		callback(r)
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
			if (!pageReport.Nofollow && !l.NoFollow) || p.FollowNofollow {
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

		var cssURLs []*url.URL

		// htmlquery panics if htmlNode is of type html.ErroNode
		if strings.HasPrefix(strings.ToLower(pageReport.ContentType), "text/html") && htmlNode.Type != html.ErrorNode {
			// Check preload links to add the urls to the crawler's queue.
			preload, err := htmlquery.QueryAll(htmlNode, "//head/link[@rel=\"preload\"]/@href")
			if err != nil {
				log.Printf("error getting preload links %v %v", preload, err)
			}

			for _, preloadLink := range preload {
				pl, err := urlutils.AbsoluteURL(htmlquery.SelectAttr(preloadLink, "href"), htmlNode, pageReport.ParsedURL)
				if err != nil {
					log.Printf("error getting preload link href %s %v", pl.String(), err)
					continue
				}

				err = c.AddRequest(&crawler.RequestMessage{URL: pl, IgnoreDomain: true, Data: requestData})
				if errors.Is(err, crawler.ErrBlockedByRobotstxt) {
					s.saveBlockedPageReport(pl, crawl)
					crawl.BlockedByRobotstxt++
				}
			}

			// extract urls from style elements
			styleTags, err := htmlquery.QueryAll(htmlNode, "//style")
			if err != nil {
				log.Printf("error getting style elements %v", err)
			}

			for _, st := range styleTags {
				cssURLs = append(cssURLs, s.ExtractURLsFromCSS(htmlquery.InnerText(st))...)
			}

			// Extract urls from inline css
			inlineStyleElements, err := htmlquery.QueryAll(htmlNode, "//*[@style]")
			if err != nil {
				log.Printf("error getting elements with style attribute: %v", err)
			}

			for _, inlineStyleElement := range inlineStyleElements {
				for _, attr := range inlineStyleElement.Attr {
					if attr.Key == "style" {
						cssURLs = append(cssURLs, s.ExtractURLsFromCSS(attr.Val)...)
					}
				}
			}
		}

		// Extract URLs from the css files
		if strings.HasPrefix(strings.ToLower(pageReport.ContentType), "text/css") {
			body, err := io.ReadAll(r.Response.Body)
			if err != nil {
				log.Printf("failed to read response body: %v", err)
			}
			cssURLs = append(cssURLs, s.ExtractURLsFromCSS(string(body))...)
		}

		// Add the extracted urls to the crawler's queue
		for _, u := range cssURLs {
			u = pageReport.ParsedURL.ResolveReference(u)

			if u.Scheme != "https" && u.Scheme != "http" {
				continue
			}

			err = c.AddRequest(&crawler.RequestMessage{URL: u, IgnoreDomain: true, Data: requestData})
			if errors.Is(err, crawler.ErrBlockedByRobotstxt) {
				s.saveBlockedPageReport(u, crawl)
				crawl.BlockedByRobotstxt++
			}
		}

		// Check the external links if the project is set to do so.
		if p.CheckExternalLinks {
			s.checkExternalLinks(c.Client, pageReport)
		}

		// Save the pageReport if it hasn't the noindex attribute or if the project
		// is set to include the noindexable URLs.
		// If the pageReport is saved correctly create the page issues, otherwise
		// log the error.
		if !pageReport.Noindex || p.IncludeNoindex {
			pageReport, err = s.repository.SavePageReport(pageReport, crawl.Id)
			if err == nil {
				headers := make(http.Header)
				if r.Response != nil {
					headers = r.Response.Header
				}
				s.reportManager.CreatePageIssues(pageReport, htmlNode, &headers, crawl)
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
	// Check if the response caused an error and save a pageReport.
	if r.Error != nil {
		log.Printf("responseMessage error: %v", r.Error)

		return &models.PageReport{
			Timeout:   true,
			URL:       r.URL.String(),
			ParsedURL: r.URL,
		}, &html.Node{}, nil
	}

	// Create a new PageReport from the response. If there's a context.DeadlineExceeded
	// error save a pageReport with a timeout.
	pageReport, htmlNode, err := NewFromHTTPResponse(r.Response)
	if err != nil {
		log.Printf("pageReport error: %v", err)

		pageReport.URL = r.URL.String()
		pageReport.ParsedURL = r.URL

		if _, ok := err.(net.Error); ok {
			pageReport.Timeout = true
		}

		if _, ok := err.(*url.Error); ok {
			pageReport.Timeout = true
		}

		if errors.Is(err, context.DeadlineExceeded) {
			pageReport.Timeout = true
		}

		if !pageReport.Timeout {
			return nil, nil, err
		}
	}

	return pageReport, htmlNode, nil
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

	_, err := s.repository.SavePageReport(pageReport, crawl.Id)
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
		if link.NoFollow || pageReport.Nofollow {
			crawl.InternalNoFollowLinks++
		} else {
			crawl.InternalFollowLinks++
		}
	}

	for _, link := range pageReport.ExternalLinks {
		if link.NoFollow || pageReport.Nofollow {
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
func (s *CrawlerHandler) checkExternalLinks(client crawler.Client, pageReport *models.PageReport) {
	for n, l := range pageReport.ExternalLinks {
		status, ok := s.externalLinksStatus[l.URL]
		if ok {
			pageReport.ExternalLinks[n].StatusCode = status
			continue
		}

		statusCode := -1
		res, err := client.Head(l.URL)
		if err == nil {
			statusCode = res.Response.StatusCode
		}

		s.externalLinksStatus[l.URL] = statusCode
		pageReport.ExternalLinks[n].StatusCode = statusCode
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

	resources = append(resources, p.Scripts...)
	resources = append(resources, p.Styles...)
	resources = append(resources, p.Audios...)

	for _, l := range p.Images {
		resources = append(resources, l.URL)
	}

	for _, l := range p.Videos {
		resources = append(resources, l.URL)
		if l.Poster != "" {
			resources = append(resources, l.Poster)
		}
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

// ExtractURLsFromCSS accepts a css string and returns a slice with all the urls
// it finds in it.
func (s *CrawlerHandler) ExtractURLsFromCSS(cssContent string) []*url.URL {
	urls := []*url.URL{}

	urlRegex := regexp.MustCompile(`url\((.*?)\)`)
	matches := urlRegex.FindAllStringSubmatch(cssContent, -1)

	for _, match := range matches {
		if len(match) > 1 {
			urlStr := match[1]
			urlStr = strings.Trim(urlStr, "'\"")
			if u, err := url.Parse(urlStr); err == nil {
				urls = append(urls, u)
			} else {
				log.Printf("Failed to parse URL: %s\n", urlStr)
			}
		}
	}

	return urls
}
