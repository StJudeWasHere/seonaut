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
type Config struct {
	Agent string `mapstructure:"agent"`
}

type Storage interface {
	SaveCrawl(project.Project) int64
	SavePageReport(*PageReport, int64)
	SaveEndCrawl(int64, time.Time, int)
	DeletePreviousCrawl(int)
}

type Service struct {
	store  Storage
	config *Config
}

func NewService(s Storage, c *Config) *Service {
	return &Service{
		store:  s,
		config: c,
	}
}

func (s *Service) StartCrawler(p project.Project) int {
	var totalURLs int

	log.Printf("Crawling %s\n", p.URL)
	start := time.Now()

	u, err := url.Parse(p.URL)
	if err != nil {
		log.Printf("startCrawler: %s %v\n", p.URL, err)
		return 0
	}

	c := NewCrawler(
		u,
		s.config.Agent,
		MaxPageReports,
		p.IgnoreRobotsTxt,
		p.FollowNofollow,
	)

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
