package services

import (
	"errors"
	"time"

	"github.com/stjudewashere/seonaut/internal/models"
)

const (
	Critical = iota + 1
	Alert
	Warning
)

type (
	IssueServiceStorage interface {
		GetNumberOfPagesForIssues(int64, string) int
		FindPageReportIssues(int64, int, string) []models.PageReport
		FindIssuesByPriority(int64, int) []models.IssueGroup
		SaveIssuesCount(int64, int, int, int)
		SaveEndIssues(int64, time.Time)
	}

	IssueService struct {
		store IssueServiceStorage
	}
)

func NewIssueService(s IssueServiceStorage) *IssueService {
	return &IssueService{store: s}
}

// GetIssuesCount returns an IssueCount with the number of issues by type.
func (s *IssueService) GetIssuesCount(crawlID int64) *models.IssueCount {
	return &models.IssueCount{
		CriticalIssues: s.store.FindIssuesByPriority(crawlID, Critical),
		AlertIssues:    s.store.FindIssuesByPriority(crawlID, Alert),
		WarningIssues:  s.store.FindIssuesByPriority(crawlID, Warning),
	}
}

// SaveCrawlIssuesCount stores the issue count in the storage.
func (s *IssueService) SaveCrawlIssuesCount(crawl *models.Crawl) {
	s.store.SaveEndIssues(crawl.Id, time.Now())
	ic := &models.IssueCount{
		CriticalIssues: s.store.FindIssuesByPriority(crawl.Id, Critical),
		AlertIssues:    s.store.FindIssuesByPriority(crawl.Id, Alert),
		WarningIssues:  s.store.FindIssuesByPriority(crawl.Id, Warning),
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
}

// Returns a PaginatorView with the corresponding page reports.
func (s *IssueService) GetPaginatedReportsByIssue(crawlId int64, currentPage int, issueId string) (models.PaginatorView, error) {
	paginator := models.Paginator{
		TotalPages:  s.store.GetNumberOfPagesForIssues(crawlId, issueId),
		CurrentPage: currentPage,
	}

	if currentPage < 1 || currentPage > paginator.TotalPages {
		return models.PaginatorView{}, errors.New("page out of bounds")
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
