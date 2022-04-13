package crawler

import (
	"database/sql"
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
	SaveCrawl(project.Project) (*Crawl, error)
	SavePageReport(*PageReport, int64)
	SaveEndCrawl(*Crawl) (*Crawl, error)
	DeletePreviousCrawl(int)
	GetLastCrawls(project.Project) []Crawl
}

type Crawl struct {
	Id          int64
	ProjectId   int
	URL         string
	Start       time.Time
	End         sql.NullTime
	TotalIssues int
	TotalURLs   int
	IssuesEnd   sql.NullTime
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

// StartCrawler creates a new crawler and crawls the project's URL
func (s *Service) StartCrawler(p project.Project) (*Crawl, error) {
	u, err := url.Parse(p.URL)
	if err != nil {
		return nil, err
	}

	c := NewCrawler(
		u,
		s.config.Agent,
		MaxPageReports,
		p.IgnoreRobotsTxt,
		p.FollowNofollow,
		p.IncludeNoindex,
	)

	crawl, err := s.store.SaveCrawl(p)
	if err != nil {
		return nil, err
	}

	pageReport := make(chan PageReport)
	go c.Crawl(pageReport)

	for r := range pageReport {
		crawl.TotalURLs++
		s.store.SavePageReport(&r, crawl.Id)
	}

	crawl, err = s.store.SaveEndCrawl(crawl)
	if err != nil {
		return nil, err
	}

	go func() {
		s.store.DeletePreviousCrawl(p.Id)
	}()

	return crawl, nil
}

func (s *Service) GetLastCrawls(p project.Project) []Crawl {
	return s.store.GetLastCrawls(p)
}
