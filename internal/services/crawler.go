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
	CrawlLimit      = 20000 // Max number of page reports that will be created
	LastCrawlsLimit = 5     // Max number returned by GetLastCrawls
)

type CrawlerServiceStorage interface {
	SaveCrawl(models.Project) (*models.Crawl, error)
	GetLastCrawl(p *models.Project) models.Crawl
	GetLastCrawls(models.Project, int) []models.Crawl
	DeleteCrawlData(c *models.Crawl)

	CountIssuesByPriority(int64, int) int
	UpdateCrawl(*models.Crawl)
}

type CrawlerServicesContainer struct {
	Broker         *Broker
	ReportManager  *ReportManager
	CrawlerHandler *CrawlerHandler
	Config         *config.CrawlerConfig
}

type CrawlerService struct {
	store          CrawlerServiceStorage
	config         *config.CrawlerConfig
	broker         *Broker
	reportManager  *ReportManager
	crawlerHandler *CrawlerHandler
	crawlers       map[int64]*crawler.Crawler
	lock           *sync.RWMutex
}

func NewCrawlerService(s CrawlerServiceStorage, services CrawlerServicesContainer) *CrawlerService {
	return &CrawlerService{
		store:          s,
		broker:         services.Broker,
		config:         services.Config,
		reportManager:  services.ReportManager,
		crawlerHandler: services.CrawlerHandler,
		crawlers:       make(map[int64]*crawler.Crawler),
		lock:           &sync.RWMutex{},
	}
}

// StartCrawler creates a new crawler and crawls the project's URL.
// It adds a new crawler for the project, it returns an error if there's one already
// running or if there's an error creating it.
// Finally the previous crawl's data is removed and the crawl is returned.
func (s *CrawlerService) StartCrawler(p models.Project) error {
	previousCrawl := s.store.GetLastCrawl(&p)
	crawl, err := s.store.SaveCrawl(p)
	if err != nil {
		return err
	}

	u, err := url.Parse(p.URL)
	if err != nil {
		return err
	}

	if u.Path == "" {
		u.Path = "/"
	}

	c, err := s.addCrawler(u, &p)
	if err != nil {
		return err
	}

	c.OnResponse(s.crawlerHandler.responseCallback(crawl, &p, c))

	go func() {
		log.Printf("Crawling %s...", p.URL)
		c.AddRequest(&crawler.RequestMessage{URL: u, Data: crawlerData{}})

		// Calling Start() initiates the website crawling process and
		// blocks execution until the crawling is complete.
		c.Start()

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

		s.removeCrawler(&p)
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

// StopCrawler stops a crawler. If the crawler does not exsit it will just return.
func (s *CrawlerService) StopCrawler(p models.Project) {
	s.lock.Lock()
	defer s.lock.Unlock()

	crawler, ok := s.crawlers[p.Id]
	if !ok {
		return
	}

	crawler.Stop()
}

// AddCrawler creates a new project crawler and adds it to the crawlers map. It returns the crawler
// on success otherwise it returns an error indicating the crawler already exists or there was an
// error creating it.
func (s *CrawlerService) addCrawler(u *url.URL, p *models.Project) (*crawler.Crawler, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.crawlers[p.Id]; ok {
		return nil, errors.New("project is already being crawled")
	}

	options := &crawler.Options{
		CrawlLimit:      CrawlLimit,
		IgnoreRobotsTxt: p.IgnoreRobotsTxt,
		FollowNofollow:  p.FollowNofollow,
		IncludeNoindex:  p.IncludeNoindex,
		UserAgent:       s.config.Agent,
		CrawlSitemap:    p.CrawlSitemap,
		AllowSubdomains: p.AllowSubdomains,
		AuthUser:        p.AuthUser,
		AuthPass:        p.AuthPass,
	}

	// Creates a new crawler with the crawler's response handler.
	s.crawlers[p.Id] = crawler.NewCrawler(u, options)

	return s.crawlers[p.Id], nil
}

// RemoveCrawler removes a project's crawler from the crawlers map.
func (s *CrawlerService) removeCrawler(p *models.Project) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.crawlers, p.Id)
}
