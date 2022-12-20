package issue

import (
	"errors"

	"github.com/stjudewashere/seonaut/internal/pagereport"
)

const (
	Critical = iota + 1
	Alert
	Warning
)

type IssueStore interface {
	CountByMediaType(int64) CountList
	CountByStatusCode(int64) CountList
	CountByCanonical(int64) int
	CountImagesAlt(int64) *AltCount
	CountScheme(int64) *SchemeCount
	CountByNonCanonical(int64) int
	GetNumberOfPagesForIssues(int64, string) int
	FindPageReportIssues(int64, int, string) []pagereport.PageReport
	FindIssuesByPriority(int64, int) []IssueGroup
	SaveIssuesCount(int64, int, int, int)
}

type Issue struct {
	PageReportId int
	CrawlId      int64
	ErrorType    int
}

type Service struct {
	store IssueStore
}

type IssueGroup struct {
	ErrorType string
	Priority  int
	Count     int
}

type CanonicalCount struct {
	Canonical    int
	NonCanonical int
}

type AltCount struct {
	Alt    int
	NonAlt int
}

type SchemeCount struct {
	HTTP  int
	HTTPS int
}

type IssueCount struct {
	MediaCount     CountList
	StatusCount    CountList
	CriticalIssues []IssueGroup
	AlertIssues    []IssueGroup
	WarningIssues  []IssueGroup
}

type Paginator struct {
	CurrentPage  int
	NextPage     int
	PreviousPage int
	TotalPages   int
}

type PaginatorView struct {
	Paginator   Paginator
	PageReports []pagereport.PageReport
}

func NewService(s IssueStore) *Service {
	return &Service{
		store: s,
	}
}

func (s *Service) GetIssuesCount(crawlID int64) *IssueCount {
	c := &IssueCount{
		MediaCount:     s.store.CountByMediaType(crawlID),
		StatusCount:    s.store.CountByStatusCode(crawlID),
		CriticalIssues: s.store.FindIssuesByPriority(crawlID, Critical),
		AlertIssues:    s.store.FindIssuesByPriority(crawlID, Alert),
		WarningIssues:  s.store.FindIssuesByPriority(crawlID, Warning),
	}

	return c
}

func (s *Service) SaveCrawlIssuesCount(crawlID int64) {
	criticalIssues := s.store.FindIssuesByPriority(crawlID, Critical)
	alertIssues := s.store.FindIssuesByPriority(crawlID, Alert)
	warningIssues := s.store.FindIssuesByPriority(crawlID, Warning)

	var critical, alert, warning int

	for _, v := range criticalIssues {
		critical += v.Count
	}

	for _, v := range alertIssues {
		alert += v.Count
	}

	for _, v := range warningIssues {
		warning += v.Count
	}

	s.store.SaveIssuesCount(crawlID, critical, alert, warning)
}

func (s *Service) GetPaginatedReportsByIssue(crawlId int64, currentPage int, issueId string) (PaginatorView, error) {
	paginator := Paginator{
		TotalPages:  s.store.GetNumberOfPagesForIssues(crawlId, issueId),
		CurrentPage: currentPage,
	}

	if currentPage < 1 || currentPage > paginator.TotalPages {
		return PaginatorView{}, errors.New("Page out of bounds")
	}

	if currentPage < paginator.TotalPages {
		paginator.NextPage = currentPage + 1
	}

	if currentPage > 1 {
		paginator.PreviousPage = currentPage - 1
	}

	paginatorView := PaginatorView{
		Paginator:   paginator,
		PageReports: s.store.FindPageReportIssues(crawlId, currentPage, issueId),
	}

	return paginatorView, nil
}

func (s *Service) GetCanonicalCount(crawlId int64) *CanonicalCount {
	return &CanonicalCount{
		Canonical:    s.store.CountByCanonical(crawlId),
		NonCanonical: s.store.CountByNonCanonical(crawlId),
	}
}

func (s *Service) GetImageAltCount(crawlId int64) *AltCount {
	return s.store.CountImagesAlt(crawlId)
}

func (s *Service) GetSchemeCount(crawlId int64) *SchemeCount {
	return s.store.CountScheme(crawlId)
}
