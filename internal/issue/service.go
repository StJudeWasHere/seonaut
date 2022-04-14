package issue

import (
	"errors"

	"github.com/stjudewashere/seonaut/internal/crawler"
)

const (
	Critical = iota + 1
	Alert
	Warning
)

type IssueStore interface {
	CountByMediaType(int64) CountList
	CountByStatusCode(int64) CountList
	GetNumberOfPagesForIssues(int64, string) int
	FindPageReportIssues(int64, int, string) []crawler.PageReport
	CountByFollowLinks(int64) CountList
	CountByFollowExternalLinks(int64) CountList
	FindIssuesByPriority(int64, int) []IssueGroup
}

type Issue struct {
	PageReportId int
	ErrorType    int
}

type IssueService struct {
	store IssueStore
}

type IssueGroup struct {
	ErrorType string
	Priority  int
	Count     int
}

type IssueCount struct {
	Critical       int
	Alert          int
	Warning        int
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
	PageReports []crawler.PageReport
}

type LinksCount struct {
	Internal CountList
	External CountList
}

func NewService(s IssueStore) *IssueService {
	return &IssueService{
		store: s,
	}
}

func (s *IssueService) GetIssuesCount(crawlID int64) *IssueCount {
	c := &IssueCount{
		MediaCount:     s.store.CountByMediaType(crawlID),
		StatusCount:    s.store.CountByStatusCode(crawlID),
		CriticalIssues: s.store.FindIssuesByPriority(crawlID, Critical),
		AlertIssues:    s.store.FindIssuesByPriority(crawlID, Alert),
		WarningIssues:  s.store.FindIssuesByPriority(crawlID, Warning),
	}

	for _, v := range c.CriticalIssues {
		c.Critical += v.Count
	}

	for _, v := range c.AlertIssues {
		c.Alert += v.Count
	}

	for _, v := range c.WarningIssues {
		c.Warning += v.Count
	}

	return c
}

func (s *IssueService) GetPaginatedReportsByIssue(crawlId int64, currentPage int, issueId string) (PaginatorView, error) {
	paginator := Paginator{
		TotalPages: s.store.GetNumberOfPagesForIssues(crawlId, issueId),
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

func (s *IssueService) GetLinksCount(crawlId int64) *LinksCount {
	l := &LinksCount{
		Internal: s.store.CountByFollowLinks(crawlId),
		External: s.store.CountByFollowExternalLinks(crawlId),
	}

	return l
}
