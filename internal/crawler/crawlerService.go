package crawler

import (
	"log"
	"net/url"
	"time"

	"github.com/stjudewashere/seonaut/internal/project"
)

const (
	// Max number of page reports that will be created
	MaxPageReports = 10000
)

// CrawlerConfig stores the configuration for the crawler.
// It is loaded from the config package.
type CrawlerConfig struct {
	Agent string `mapstructure:"agent"`
}

type CrawlerStore interface {
	SaveCrawl(project.Project) int64
	SavePageReport(*PageReport, int64)
	SaveEndCrawl(int64, time.Time, int)
	DeletePreviousCrawl(int)
}

type CrawlerService struct {
	store  CrawlerStore
	config *CrawlerConfig
}

func NewService(s CrawlerStore, c *CrawlerConfig) *CrawlerService {
	return &CrawlerService{
		store:  s,
		config: c,
	}
}

func (s *CrawlerService) StartCrawler(p project.Project) int {
	var totalURLs int

	log.Printf("Crawling %s\n", p.URL)
	start := time.Now()

	u, err := url.Parse(p.URL)
	if err != nil {
		log.Printf("startCrawler: %s %v\n", p.URL, err)
		return 0
	}

	c := &Crawler{
		URL:             u,
		MaxPageReports:  MaxPageReports,
		IgnoreRobotsTxt: p.IgnoreRobotsTxt,
		FollowNofollow:  p.FollowNofollow,
		UserAgent:       s.config.Agent,
	}

	cid := s.store.SaveCrawl(p)

	pageReport := make(chan PageReport)
	go c.Crawl(pageReport)

	for r := range pageReport {
		totalURLs++
		s.store.SavePageReport(&r, cid)
	}

	s.store.SaveEndCrawl(cid, time.Now(), totalURLs)
	log.Printf("Done crawling %s in %s\n", p.URL, time.Since(start))
	log.Printf("%d pages crawled.\n", totalURLs)

	go func() {
		log.Printf("Deleting previous crawl data for %s\n", p.URL)
		s.store.DeletePreviousCrawl(p.Id)
		log.Printf("Deleted previous crawl done for %s\n", p.URL)
	}()

	return int(cid)
}
