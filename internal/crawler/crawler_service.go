package crawler

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/stjudewashere/seonaut/internal/cache_manager"
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/pubsub"
	"github.com/stjudewashere/seonaut/internal/report_manager"
)

const (
	// Max number of page reports that will be created
	MaxPageReports = 20000

	// Max number returned by GetLastCrawls
	LastCrawlsLimit = 5
)

// CrawlerConfig stores the configuration for the crawler.
// It is loaded from the config package.
type Config struct {
	Agent string `mapstructure:"agent"`
}

type Storage interface {
	SaveCrawl(models.Project) (*models.Crawl, error)
	SavePageReport(*models.PageReport, int64) (*models.PageReport, error)
	SaveEndCrawl(*models.Crawl) (*models.Crawl, error)
	GetLastCrawls(models.Project, int) []models.Crawl
	GetPreviousCrawl(*models.Project) (*models.Crawl, error)
	DeleteCrawlData(c *models.Crawl)
}
type Service struct {
	store         Storage
	broker        *pubsub.Broker
	config        *Config
	cacheManager  *cache_manager.CacheManager
	reportManager *report_manager.ReportManager
	crawlers      map[int64]*Crawler
	lock          *sync.RWMutex
}

func NewService(s Storage, broker *pubsub.Broker, c *Config, cm *cache_manager.CacheManager, rm *report_manager.ReportManager) *Service {
	return &Service{
		store:         s,
		broker:        broker,
		config:        c,
		cacheManager:  cm,
		reportManager: rm,
		crawlers:      make(map[int64]*Crawler),
		lock:          &sync.RWMutex{},
	}
}

// StartCrawler creates a new crawler and crawls the project's URL
func (s *Service) StartCrawler(p models.Project) (*models.Crawl, error) {
	u, err := url.Parse(p.URL)
	if err != nil {
		return nil, err
	}

	if u.Path == "" {
		u.Path = "/"
	}

	options := &Options{
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

	c := NewCrawler(u, options)

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

		r.PageReport, err = s.store.SavePageReport(r.PageReport, crawl.Id)
		if err != nil {
			log.Printf("SavePageReport: %v\n", err)
			continue
		}

		s.reportManager.CreatePageIssues(r.PageReport, r.HtmlNode, r.Header, crawl)

		s.broker.Publish(fmt.Sprintf("crawl-%d", p.Id), &pubsub.Message{Name: "PageReport", Data: r})
	}

	crawl.RobotstxtExists = c.RobotstxtExists()
	crawl.SitemapExists = c.SitemapExists()
	crawl.SitemapIsBlocked = c.SitemapIsBlocked()

	crawl, err = s.store.SaveEndCrawl(crawl)
	if err != nil {
		return nil, err
	}

	go func() {
		previous, err := s.store.GetPreviousCrawl(&p)
		if err != nil {
			log.Printf("Crawler: PreviousCrawl: %v\n", err)
			return
		}

		s.store.DeleteCrawlData(previous)
		s.cacheManager.RemoveCrawlCache(previous)
	}()

	return crawl, nil
}

// Get a slice with 'LastCrawlsLimit' number of the crawls
func (s *Service) GetLastCrawls(p models.Project) []models.Crawl {
	crawls := s.store.GetLastCrawls(p, LastCrawlsLimit)

	for len(crawls) < LastCrawlsLimit {
		crawls = append(crawls, models.Crawl{Start: time.Now()})
	}

	return crawls
}

// Get the crawler from the crawlers map and stop it.
// In case the crawler is not running it just returns.
func (s *Service) StopCrawler(p models.Project) {
	crawler, ok := s.crawlers[p.Id]
	if !ok {
		return
	}

	crawler.Stop()
}
