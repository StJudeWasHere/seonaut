package services

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/stjudewashere/seonaut/internal/config"
	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/models"
)

const (
	MaxPageReports  = 20000 // Max number of page reports that will be created
	LastCrawlsLimit = 5     // Max number returned by GetLastCrawls
)

type (
	CrawlerPageReportStorage interface {
		SavePageReport(*models.PageReport, int64) (*models.PageReport, error)
	}

	CrawlerServiceStorage interface {
		SaveCrawl(models.Project) (*models.Crawl, error)
		SaveEndCrawl(*models.Crawl) (*models.Crawl, error)
		GetLastCrawls(models.Project, int) []models.Crawl
		GetPreviousCrawl(*models.Project) (*models.Crawl, error)
		DeleteCrawlData(c *models.Crawl)
	}

	Services struct {
		Broker        *Broker
		ReportManager *ReportManager
		IssueService  *IssueService
	}

	CrawlerService struct {
		store         CrawlerServiceStorage
		pstore        CrawlerPageReportStorage
		broker        *Broker
		config        *config.CrawlerConfig
		reportManager *ReportManager
		issueService  *IssueService
		crawlers      map[int64]*crawler.Crawler
		lock          *sync.RWMutex
	}
)

func NewCrawlerService(s CrawlerServiceStorage, ps CrawlerPageReportStorage, c *config.CrawlerConfig, services Services) *CrawlerService {
	return &CrawlerService{
		store:         s,
		pstore:        ps,
		broker:        services.Broker,
		config:        c,
		reportManager: services.ReportManager,
		issueService:  services.IssueService,
		crawlers:      make(map[int64]*crawler.Crawler),
		lock:          &sync.RWMutex{},
	}
}

// StartCrawler creates a new crawler and crawls the project's URL
func (s *CrawlerService) StartCrawler(p models.Project) (*models.Crawl, error) {
	u, err := url.Parse(p.URL)
	if err != nil {
		return nil, err
	}

	if u.Path == "" {
		u.Path = "/"
	}

	options := &crawler.Options{
		MaxPageReports:     MaxPageReports,
		IgnoreRobotsTxt:    p.IgnoreRobotsTxt,
		FollowNofollow:     p.FollowNofollow,
		IncludeNoindex:     p.IncludeNoindex,
		UserAgent:          s.config.Agent,
		CrawlSitemap:       p.CrawlSitemap,
		AllowSubdomains:    p.AllowSubdomains,
		BasicAuth:          p.BasicAuth,
		AuthUser:           p.AuthUser,
		AuthPass:           p.AuthPass,
		CheckExternalLinks: p.CheckExternalLinks,
	}

	crawl, err := s.store.SaveCrawl(p)
	if err != nil {
		return nil, err
	}

	if _, ok := s.crawlers[p.Id]; ok {
		return nil, errors.New("project is already being crawled")
	}

	c := crawler.NewCrawler(u, options)

	s.lock.Lock()
	s.crawlers[p.Id] = c
	s.lock.Unlock()

	defer func(id int64) {
		s.lock.Lock()
		delete(s.crawlers, id)
		s.lock.Unlock()
	}(p.Id)

	for r := range c.Stream() {
		// URLs are added to the TotalURLs count if they are not blocked
		// by the robots.txt and they are indexable.
		// Otherwise they are added to the BlockedByRobotstxt or Noindex count.
		if r.PageReport.BlockedByRobotstxt {
			crawl.BlockedByRobotstxt++
		} else if r.PageReport.Noindex {
			crawl.Noindex++
		}

		crawl.TotalURLs++

		// Count total internal follow and nofollow links.
		for _, l := range r.PageReport.Links {
			if l.NoFollow {
				crawl.InternalNoFollowLinks++
			} else {
				crawl.InternalFollowLinks++
			}
		}

		// Count total external follow, nofollow, sponsored and UGC links.
		for _, l := range r.PageReport.ExternalLinks {
			if l.NoFollow {
				crawl.ExternalNoFollowLinks++
			} else {
				crawl.ExternalFollowLinks++
			}

			if l.Sponsored {
				crawl.SponsoredLinks++
			}

			if l.UGC {
				crawl.UGCLinks++
			}
		}

		r.PageReport, err = s.pstore.SavePageReport(r.PageReport, crawl.Id)
		if err != nil {
			log.Printf("SavePageReport: %v\n", err)
			continue
		}

		s.reportManager.CreatePageIssues(r.PageReport, r.HtmlNode, r.Header, crawl)

		s.broker.Publish(fmt.Sprintf("crawl-%d", p.Id), &models.Message{Name: "PageReport", Data: r})
	}

	crawl.RobotstxtExists = c.RobotstxtExists()
	crawl.SitemapExists = c.SitemapExists()
	crawl.SitemapIsBlocked = c.SitemapIsBlocked()

	crawl, err = s.store.SaveEndCrawl(crawl)
	if err != nil {
		return nil, err
	}

	s.broker.Publish(fmt.Sprintf("crawl-%d", p.Id), &models.Message{Name: "IssuesInit"})
	s.reportManager.CreateMultipageIssues(crawl)

	s.issueService.SaveCrawlIssuesCount(crawl)
	s.broker.Publish(fmt.Sprintf("crawl-%d", p.Id), &models.Message{Name: "CrawlEnd", Data: crawl.TotalURLs})

	go func() {
		previous, err := s.store.GetPreviousCrawl(&p)
		if err != nil {
			log.Printf("Crawler: PreviousCrawl: %v\n", err)
			return
		}

		s.store.DeleteCrawlData(previous)
	}()

	return crawl, nil
}

// Get a slice with 'LastCrawlsLimit' number of the crawls
func (s *CrawlerService) GetLastCrawls(p models.Project) []models.Crawl {
	crawls := s.store.GetLastCrawls(p, LastCrawlsLimit)

	for len(crawls) < LastCrawlsLimit {
		crawls = append(crawls, models.Crawl{Start: time.Now()})
	}

	return crawls
}

// Get the crawler from the crawlers map and stop it.
// In case the crawler is not running it just returns.
func (s *CrawlerService) StopCrawler(p models.Project) {
	crawler, ok := s.crawlers[p.Id]
	if !ok {
		return
	}

	crawler.Stop()
}
