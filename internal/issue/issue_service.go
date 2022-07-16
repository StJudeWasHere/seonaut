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
	CountByCanonical(int64) int
	CountImagesAlt(int64) *AltCount
	CountScheme(int64) *SchemeCount
	CountByNonCanonical(int64) int
	CountSponsoredLinks(int64) int
	CountUGCLinks(int64) int
	GetNumberOfPagesForIssues(int64, string) int
	FindPageReportIssues(int64, int, string) []crawler.PageReport
	CountByFollowLinks(int64) CountList
	CountByFollowExternalLinks(int64) CountList
	FindIssuesByPriority(int64, int) []IssueGroup
	SaveIssuesCount(int64, int, int, int)
}

type Issue struct {
	PageReportId int
	CrawlId      int64
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
	Internal      CountList
	External      CountList
	Total         int
	TotalInternal int
	TotalExternal int
	Sponsored     int
	UGC           int
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

func (s *IssueService) SaveCrawlIssuesCount(crawlID int64) {
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

func (s *IssueService) GetPaginatedReportsByIssue(crawlId int64, currentPage int, issueId string) (PaginatorView, error) {
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

func (s *IssueService) GetLinksCount(crawlId int64) *LinksCount {
	l := &LinksCount{
		Internal:  s.store.CountByFollowLinks(crawlId),
		External:  s.store.CountByFollowExternalLinks(crawlId),
		Sponsored: s.store.CountSponsoredLinks(crawlId),
		UGC:       s.store.CountUGCLinks(crawlId),
	}

	for _, v := range l.Internal {
		l.Total += v.Value
		l.TotalInternal += v.Value
	}

	for _, v := range l.External {
		l.Total += v.Value
		l.TotalExternal += v.Value
	}

	return l
}

func (s *IssueService) GetCanonicalCount(crawlId int64) *CanonicalCount {
	return &CanonicalCount{
		Canonical:    s.store.CountByCanonical(crawlId),
		NonCanonical: s.store.CountByNonCanonical(crawlId),
	}
}

func (s *IssueService) GetImageAltCount(crawlId int64) *AltCount {
	return s.store.CountImagesAlt(crawlId)
}

func (s *IssueService) GetSchemeCount(crawlId int64) *SchemeCount {
	return s.store.CountScheme(crawlId)
}
