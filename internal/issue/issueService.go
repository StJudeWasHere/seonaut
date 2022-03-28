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

const (
	Error30x = iota + 1
	Error40x
	Error50x
	ErrorDuplicatedTitle
	ErrorDuplicatedDescription
	ErrorEmptyTitle
	ErrorShortTitle
	ErrorLongTitle
	ErrorEmptyDescription
	ErrorShortDescription
	ErrorLongDescription
	ErrorLittleContent
	ErrorImagesWithNoAlt
	ErrorRedirectChain
	ErrorNoH1
	ErrorNoLang
	ErrorHTTPLinks
	ErrorHreflangsReturnLink
	ErrorTooManyLinks
	ErrorInternalNoFollow
	ErrorExternalWithoutNoFollow
	ErrorCanonicalizedToNonCanonical
	ErrorRedirectLoop
	ErrorNotValidHeadings
)

type IssueStore interface {
	FindIssues(int) map[string]IssueGroup
	CountByMediaType(int) CountList
	CountByStatusCode(int) CountList
	GetNumberOfPagesForIssues(int, string) int
	FindPageReportIssues(int, int, string) []crawler.PageReport
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
	Groups      map[string]IssueGroup
	Critical    int
	Alert       int
	Warning     int
	MediaCount  CountList
	StatusCount CountList
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

func NewService(s IssueStore) *IssueService {
	return &IssueService{
		store: s,
	}
}

func (s *IssueService) GetIssuesCount(crawlID int) *IssueCount {
	c := &IssueCount{
		Groups:      s.store.FindIssues(crawlID),
		MediaCount:  s.store.CountByMediaType(crawlID),
		StatusCount: s.store.CountByStatusCode(crawlID),
	}

	for _, v := range c.Groups {
		switch v.Priority {
		case Critical:
			c.Critical += v.Count
		case Alert:
			c.Alert += v.Count
		case Warning:
			c.Warning += v.Count
		}
	}

	return c
}

func (s *IssueService) GetPaginatedReportsByIssue(crawlId, currentPage int, issueId string) (PaginatorView, error) {
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
