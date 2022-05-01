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

	// Max number returned by GetLastCrawls
	LastCrawlsLimit = 5
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
	GetLastCrawls(project.Project, int) []Crawl
}

type Crawl struct {
	Id                 int64
	ProjectId          int
	URL                string
	Start              time.Time
	End                sql.NullTime
	TotalIssues        int
	TotalURLs          int
	IssuesEnd          sql.NullTime
	CriticalIssues     int
	WarningIssues      int
	NoticeIssues       int
	BlockedByRobotstxt int
	Noindex            int
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
		if r.BlockedByRobotstxt {
			crawl.BlockedByRobotstxt++
		} else if r.Noindex {
			crawl.Noindex++
		} else {
			crawl.TotalURLs++
		}

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

// Get a slice with 'LastCrawlsLimit' number of the crawls
func (s *Service) GetLastCrawls(p project.Project) []Crawl {
	crawls := s.store.GetLastCrawls(p, LastCrawlsLimit)

	for len(crawls) < LastCrawlsLimit {
		crawls = append(crawls, Crawl{Start: time.Now()})
	}

	return crawls
}
