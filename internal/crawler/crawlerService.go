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
	consumerThreads        = 2
	storageMaxSize         = 10000
	MaxPageReports         = 300
	AdvancedMaxPageReports = 5000
	RendertronURL          = "http://127.0.0.1:3000/render/"
)

type CrawlerStore interface {
	SaveCrawl(project.Project) int64
	SavePageReport(*report.PageReport, int64)
	SaveEndCrawl(int64, time.Time, int)
	GetLastCrawl(*project.Project) Crawl
}

type CrawlerService struct {
	store CrawlerStore
}

func NewService(s CrawlerStore) *CrawlerService {
	return &CrawlerService{
		store: s,
	}
}

func (s *CrawlerService) StartCrawler(p project.Project, agent string, advanced bool, sanitizer *bluemonday.Policy) int {
	var totalURLs int
	var max int

	if advanced {
		max = AdvancedMaxPageReports
	} else {
		max = MaxPageReports
	}

	u, err := url.Parse(p.URL)
	if err != nil {
		log.Printf("startCrawler: %s %v\n", p.URL, err)
		return 0
	}

	c := &Crawler{
		URL:             u,
		MaxPageReports:  max,
		UseJS:           p.UseJS,
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
	log.Printf("%d pages crawled.\n", totalURLs)

	return int(cid)
}
