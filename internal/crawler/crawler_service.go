package crawler

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/stjudewashere/seonaut/internal/pagereport"
	"github.com/stjudewashere/seonaut/internal/project"
	"github.com/stjudewashere/seonaut/internal/pubsub"
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
	SaveCrawl(project.Project) (*Crawl, error)
	SavePageReport(*pagereport.PageReport, int64)
	SaveEndCrawl(*Crawl) (*Crawl, error)
	DeletePreviousCrawl(int64)
	GetLastCrawls(project.Project, int) []Crawl
}

type PageReportMessage struct {
	PageReport *pagereport.PageReport
	Crawled    int
	Discovered int
}

type Crawl struct {
	Id                    int64
	ProjectId             int64
	URL                   string
	Start                 time.Time
	End                   sql.NullTime
	TotalIssues           int
	TotalURLs             int
	IssuesEnd             sql.NullTime
	CriticalIssues        int
	AlertIssues           int
	WarningIssues         int
	BlockedByRobotstxt    int // URLs blocked by robots.txt
	Noindex               int // URLS with noindex attribute
	SitemapExists         bool
	RobotstxtExists       bool
	InternalFollowLinks   int
	InternalNoFollowLinks int
	ExternalFollowLinks   int
	ExternalNoFollowLinks int
	SponsoredLinks        int
	UGCLinks              int
}

type Service struct {
	store  Storage
	broker *pubsub.Broker
	config *Config
}

func NewService(s Storage, broker *pubsub.Broker, c *Config) *Service {
	return &Service{
		store:  s,
		broker: broker,
		config: c,
	}
}

// StartCrawler creates a new crawler and crawls the project's URL
func (s *Service) StartCrawler(p project.Project) (*Crawl, error) {
	u, err := url.Parse(p.URL)
	if err != nil {
		return nil, err
	}

	if u.Path == "" {
		u.Path = "/"
	}

	options := &Options{
		MaxPageReports:  MaxPageReports,
		IgnoreRobotsTxt: p.IgnoreRobotsTxt,
		FollowNofollow:  p.FollowNofollow,
		IncludeNoindex:  p.IncludeNoindex,
		UserAgent:       s.config.Agent,
		CrawlSitemap:    p.CrawlSitemap,
		AllowSubdomains: p.AllowSubdomains,
		BasicAuth:       p.BasicAuth,
		AuthUser:        p.AuthUser,
		AuthPass:        p.AuthPass,
	}

	crawl, err := s.store.SaveCrawl(p)
	if err != nil {
		return nil, err
	}

	c := NewCrawler(u, options)

	for r := range c.Stream() {
		// URLs are added to the TotalURLs count if they are not blocked
		// by the robots.txt and they are indexable.
		// Otherwise they are added to the BlockedByRobotstxt or Noindex count.
		if r.PageReport.BlockedByRobotstxt {
			crawl.BlockedByRobotstxt++
		} else if r.PageReport.Noindex {
			crawl.Noindex++
		} else {
			crawl.TotalURLs++
		}

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

		s.store.SavePageReport(r.PageReport, crawl.Id)
		s.broker.Publish(fmt.Sprintf("crawl-%d", p.Id), &pubsub.Message{Name: "PageReport", Data: r})
	}

	crawl.RobotstxtExists = c.RobotstxtExists()
	crawl.SitemapExists = c.SitemapExists()

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
