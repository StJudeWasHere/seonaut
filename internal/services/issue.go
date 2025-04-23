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
	IssueServiceRepository interface {
		GetNumberOfPagesForIssues(int64, string) int
		FindPageReportIssues(int64, int, string) []models.PageReport
		FindIssuesByTypeAndPriority(int64, int) []models.IssueGroup
		FindPassedIssues(cid int64) []models.IssueGroup
	}

	IssueService struct {
		repository IssueServiceRepository
	}
)

func NewIssueService(r IssueServiceRepository) *IssueService {
	return &IssueService{repository: r}
}

// GetIssuesCount returns an IssueCount with the number of issues by type.
func (s *IssueService) GetIssuesCount(crawlID int64) *models.IssueCount {
	return &models.IssueCount{
		CriticalIssues: s.repository.FindIssuesByTypeAndPriority(crawlID, Critical),
		AlertIssues:    s.repository.FindIssuesByTypeAndPriority(crawlID, Alert),
		WarningIssues:  s.repository.FindIssuesByTypeAndPriority(crawlID, Warning),
		PassedIssues:   s.repository.FindPassedIssues(crawlID),
	}
}

// Returns a PaginatorView with the corresponding page reports.
func (s *IssueService) GetPaginatedReportsByIssue(crawlId int64, currentPage int, issueId string) (models.PaginatorView, error) {
	paginator := models.Paginator{
		TotalPages:  s.repository.GetNumberOfPagesForIssues(crawlId, issueId),
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
		PageReports: s.repository.FindPageReportIssues(crawlId, currentPage, issueId),
	}

	return paginatorView, nil
}
