package crawler

import (
	"log"
	"net/url"
	"time"

	"github.com/mnlg/lenkrr/internal/project"
	"github.com/mnlg/lenkrr/internal/report"

	"github.com/microcosm-cc/bluemonday"
)

const (
	consumerThreads = 2
	storageMaxSize  = 10000
	MaxPageReports  = 10000
)

type CrawlerStore interface {
	SaveCrawl(project.Project) int64
	SavePageReport(*report.PageReport, int64)
	SaveEndCrawl(int64, time.Time, int)
	DeletePreviousCrawl(int)
}

type CrawlerService struct {
	store CrawlerStore
}

func NewService(s CrawlerStore) *CrawlerService {
	return &CrawlerService{
		store: s,
	}
}

func (s *CrawlerService) StartCrawler(p project.Project, agent string, sanitizer *bluemonday.Policy) int {
	var totalURLs int
	var max int

	log.Printf("Crawling %s\n", p.URL)
	start := time.Now()

	max = MaxPageReports

	u, err := url.Parse(p.URL)
	if err != nil {
		log.Printf("startCrawler: %s %v\n", p.URL, err)
		return 0
	}

	c := &Crawler{
		URL:             u,
		MaxPageReports:  max,
		IgnoreRobotsTxt: p.IgnoreRobotsTxt,
		UserAgent:       agent,
		sanitizer:       sanitizer,
	}

	cid := s.store.SaveCrawl(p)

	pageReport := make(chan report.PageReport)
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
