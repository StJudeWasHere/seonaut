package services

import (
	"errors"

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
		FindIssuesByTypeAndPriority(int64, int) []models.IssueGroup
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
		CriticalIssues: s.store.FindIssuesByTypeAndPriority(crawlID, Critical),
		AlertIssues:    s.store.FindIssuesByTypeAndPriority(crawlID, Alert),
		WarningIssues:  s.store.FindIssuesByTypeAndPriority(crawlID, Warning),
	}
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
