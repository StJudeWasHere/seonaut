package issue

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/stjudewashere/seonaut/internal/models"
)

const (
	Critical = iota + 1
	Alert
	Warning
)

type Cache interface {
	Set(key string, v interface{}) error
	Get(key string, v interface{}) error
	Delete(key string) error
}

type IssueStore interface {
	GetNumberOfPagesForIssues(int64, string) int
	FindPageReportIssues(int64, int, string) []models.PageReport
	FindIssuesByPriority(int64, int) []IssueGroup
	SaveIssuesCount(int64, int, int, int)
	SaveEndIssues(int64, time.Time)
}

type Service struct {
	store IssueStore
	cache Cache
}

type IssueGroup struct {
	ErrorType string
	Priority  int
	Count     int
}

type IssueCount struct {
	CriticalIssues []IssueGroup
	AlertIssues    []IssueGroup
	WarningIssues  []IssueGroup
}

func NewService(s IssueStore, c Cache) *Service {
	return &Service{
		store: s,
		cache: c,
	}
}

// GetIssuesCount returns an IssueCount with the number of issues by type.
// It checks if the data has been cached, otherwise, it creates the IssueCount and adds it to the cache.
func (s *Service) GetIssuesCount(crawlID int64) *IssueCount {
	key := fmt.Sprintf("crawl-%d", crawlID)
	v := &IssueCount{}
	err := s.cache.Get(key, v)
	if err != nil {
		v = &IssueCount{
			CriticalIssues: s.store.FindIssuesByPriority(crawlID, Critical),
			AlertIssues:    s.store.FindIssuesByPriority(crawlID, Alert),
			WarningIssues:  s.store.FindIssuesByPriority(crawlID, Warning),
		}

		if err := s.cache.Set(key, v); err != nil {
			log.Printf("GetIssuesCount: cacheSet: %v\n", err)
		}
	}

	return v
}

// SaveCrawlIssuesCount stores the issue count in the storage and adds the IssueCount to the cache.
func (s *Service) SaveCrawlIssuesCount(crawl *models.Crawl) {

	s.store.SaveEndIssues(crawl.Id, time.Now())

	key := fmt.Sprintf("crawl-%d", crawl.Id)
	ic := &IssueCount{
		CriticalIssues: s.store.FindIssuesByPriority(crawl.Id, Critical),
		AlertIssues:    s.store.FindIssuesByPriority(crawl.Id, Alert),
		WarningIssues:  s.store.FindIssuesByPriority(crawl.Id, Warning),
	}

	if err := s.cache.Set(key, ic); err != nil {
		log.Printf("GetIssuesCount: cacheSet: %v\n", err)
	}

	var critical, alert, warning int

	for _, v := range ic.CriticalIssues {
		critical += v.Count
	}

	for _, v := range ic.AlertIssues {
		alert += v.Count
	}

	for _, v := range ic.WarningIssues {
		warning += v.Count
	}

	s.store.SaveIssuesCount(crawl.Id, critical, alert, warning)
	s.BuildCrawlCache(crawl)
}

// Returns a PaginatorView with the corresponding page reports.
func (s *Service) GetPaginatedReportsByIssue(crawlId int64, currentPage int, issueId string) (models.PaginatorView, error) {
	paginator := models.Paginator{
		TotalPages:  s.store.GetNumberOfPagesForIssues(crawlId, issueId),
		CurrentPage: currentPage,
	}

	if currentPage < 1 || currentPage > paginator.TotalPages {
		return models.PaginatorView{}, errors.New("Page out of bounds")
	}

	if currentPage < paginator.TotalPages {
		paginator.NextPage = currentPage + 1
	}

	if currentPage > 1 {
		paginator.PreviousPage = currentPage - 1
	}

	paginatorView := models.PaginatorView{
		Paginator:   paginator,
		PageReports: s.store.FindPageReportIssues(crawlId, currentPage, issueId),
	}

	return paginatorView, nil
}

func (s *Service) BuildCrawlCache(crawl *models.Crawl) {
	key := fmt.Sprintf("crawl-%d", crawl.Id)
	ic := &IssueCount{
		CriticalIssues: s.store.FindIssuesByPriority(crawl.Id, Critical),
		AlertIssues:    s.store.FindIssuesByPriority(crawl.Id, Alert),
		WarningIssues:  s.store.FindIssuesByPriority(crawl.Id, Warning),
	}
	if err := s.cache.Set(key, ic); err != nil {
		log.Printf("GetIssuesCount: cacheSet: %v\n", err)
	}
}

func (s *Service) RemoveCrawlCache(crawl *models.Crawl) {
	key := fmt.Sprintf("crawl-%d", crawl.Id)
	if err := s.cache.Delete(key); err != nil {
		log.Printf("DeleteIssuesCache: %v\n", err)
	}
}
