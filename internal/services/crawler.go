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
	CrawlerServiceStorage interface {
		SaveCrawl(models.Project) (*models.Crawl, error)
		GetLastCrawl(p *models.Project) models.Crawl
		GetLastCrawls(models.Project, int) []models.Crawl
		DeleteCrawlData(c *models.Crawl)

		CountIssuesByPriority(int64, int) int
		UpdateCrawl(*models.Crawl)

		SavePageReport(*models.PageReport, int64) (*models.PageReport, error)
	}

	CrawlerServicesContainer struct {
		Broker        *Broker
		ReportManager *ReportManager
		Config        *config.CrawlerConfig
	}

	CrawlerService struct {
		store          CrawlerServiceStorage
		config         *config.CrawlerConfig
		broker         *Broker
		reportManager  *ReportManager
		crawlerManager *CrawlerManager
	}
)

func NewCrawlerService(s CrawlerServiceStorage, services CrawlerServicesContainer) *CrawlerService {
	crawlerManager := &CrawlerManager{
		config:   services.Config,
		crawlers: make(map[int64]*crawler.Crawler),
		lock:     &sync.RWMutex{},
	}

	return &CrawlerService{
		store:          s,
		broker:         services.Broker,
		config:         services.Config,
		reportManager:  services.ReportManager,
		crawlerManager: crawlerManager,
	}
}

// StartCrawler creates a new crawler and crawls the project's URL.
// It adds a new crawler for the project, it returns an error if there's one already
// running or if there's an error creating it.
// A crawl is created and it is updated with the crawler's data as urls are crawled.
// Finally the previous crawl's data is removed and the crawl is returned.
func (s *CrawlerService) StartCrawler(p models.Project) error {
	c, err := s.crawlerManager.AddCrawler(&p)
	if err != nil {
		return err
	}

	previousCrawl := s.store.GetLastCrawl(&p)
	crawl, err := s.store.SaveCrawl(p)
	if err != nil {
		return err
	}

	go func() {
		log.Printf("Crawling %s...", p.URL)
		for r := range c.Stream() {
			crawl.TotalURLs++

			if r.PageReport.BlockedByRobotstxt {
				crawl.BlockedByRobotstxt++
			}

			if r.PageReport.Noindex {
				crawl.Noindex++
			}

			for _, l := range r.PageReport.Links {
				if l.NoFollow {
					crawl.InternalNoFollowLinks++
				} else {
					crawl.InternalFollowLinks++
				}
			}

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

			r.PageReport, err = s.store.SavePageReport(r.PageReport, crawl.Id)
			if err != nil {
				log.Printf("crawler service: SavePageReport: %v\n", err)
				continue
			}

			s.reportManager.CreatePageIssues(r.PageReport, r.HtmlNode, r.Header, crawl)
			s.broker.Publish(fmt.Sprintf("crawl-%d", p.Id), &models.Message{Name: "PageReport", Data: r})
		}

		crawl.RobotstxtExists = c.RobotstxtExists()
		crawl.SitemapExists = c.SitemapExists()
		crawl.SitemapIsBlocked = c.SitemapIsBlocked()
		crawl.End = time.Now()

		s.broker.Publish(fmt.Sprintf("crawl-%d", p.Id), &models.Message{Name: "IssuesInit"})
		s.reportManager.CreateMultipageIssues(crawl)

		crawl.IssuesEnd = time.Now()
		crawl.CriticalIssues = s.store.CountIssuesByPriority(crawl.Id, Critical)
		crawl.AlertIssues = s.store.CountIssuesByPriority(crawl.Id, Alert)
		crawl.WarningIssues = s.store.CountIssuesByPriority(crawl.Id, Warning)
		crawl.TotalIssues = crawl.CriticalIssues + crawl.AlertIssues + crawl.WarningIssues

		s.store.UpdateCrawl(crawl)
		s.broker.Publish(fmt.Sprintf("crawl-%d", p.Id), &models.Message{Name: "CrawlEnd", Data: crawl.TotalURLs})
		log.Printf("Crawled %d urls in %s", crawl.TotalURLs, p.URL)

		s.crawlerManager.RemoveCrawler(&p)
		s.store.DeleteCrawlData(&previousCrawl)
	}()

	return nil
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
	s.crawlerManager.StopCrawler(&p)
}

type CrawlerManager struct {
	config   *config.CrawlerConfig
	crawlers map[int64]*crawler.Crawler
	lock     *sync.RWMutex
}

// AddCrawler creates a new project crawler and adds it to the crawlers map. It returns the crawler
// on success otherwise it returns an error indicating the crawler already exists or there was an
// error creating it.
func (s *CrawlerManager) AddCrawler(p *models.Project) (*crawler.Crawler, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.crawlers[p.Id]; ok {
		return nil, errors.New("project is already being crawled")
	}

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
	s.crawlers[p.Id] = crawler.NewCrawler(u, options)

	return s.crawlers[p.Id], nil
}

// RemoveCrawler removes a project's crawler from the crawlers map.
func (s *CrawlerManager) RemoveCrawler(p *models.Project) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.crawlers, p.Id)
}

// StopCrawler stops a crawler. If the crawler does not exsit it will just return.
func (s *CrawlerManager) StopCrawler(p *models.Project) {
	s.lock.Lock()
	defer s.lock.Unlock()

	crawler, ok := s.crawlers[p.Id]
	if !ok {
		return
	}

	crawler.Stop()
}
