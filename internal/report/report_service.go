package report

import (
	"github.com/stjudewashere/seonaut/internal/crawler"
)

type ReportStore interface {
	FindPageReportById(int) crawler.PageReport
	FindErrorTypesByPage(int, int64) []string
	FindInLinks(string, int64, int) []crawler.PageReport
	FindPageReportsRedirectingToURL(string, int64, int) []crawler.PageReport
	FindAllPageReportsByCrawlIdAndErrorType(int64, string) <-chan *crawler.PageReport
	FindAllPageReportsByCrawlId(int64) <-chan *crawler.PageReport
	FindSitemapPageReports(int64) <-chan *crawler.PageReport
	FindLinks(pageReport *crawler.PageReport, cid int64, page int) []crawler.Link
	FindExternalLinks(pageReport *crawler.PageReport, cid int64, p int) []crawler.Link

	GetNumberOfPagesForInlinks(*crawler.PageReport, int64) int
	GetNumberOfPagesForRedirecting(*crawler.PageReport, int64) int
	GetNumberOfPagesForLinks(*crawler.PageReport, int64) int
	GetNumberOfPagesForExternalLinks(pageReport *crawler.PageReport, cid int64) int
}

type ReportService struct {
	store ReportStore
}

type PageReportView struct {
	PageReport crawler.PageReport
	ErrorTypes []string
	InLinks    []crawler.PageReport
	Redirects  []crawler.PageReport
	Paginator  Paginator
}

type Paginator struct {
	CurrentPage  int
	NextPage     int
	PreviousPage int
	TotalPages   int
}

func NewService(store ReportStore) *ReportService {
	return &ReportService{
		store: store,
	}
}

// Returns a PageReportView by PageReport Id and Crawl Id
// it also loads the data specified in the tab paramater
func (s *ReportService) GetPageReport(rid int, crawlId int64, tab string, page int) *PageReportView {
	paginator := Paginator{
		CurrentPage: page,
	}

	v := &PageReportView{
		PageReport: s.store.FindPageReportById(rid),
		ErrorTypes: s.store.FindErrorTypesByPage(rid, crawlId),
	}

	switch tab {
	case "internal":
		paginator.TotalPages = s.store.GetNumberOfPagesForLinks(&v.PageReport, crawlId)
		v.PageReport.Links = s.store.FindLinks(&v.PageReport, crawlId, page)
	case "external":
		paginator.TotalPages = s.store.GetNumberOfPagesForExternalLinks(&v.PageReport, crawlId)
		v.PageReport.ExternalLinks = s.store.FindExternalLinks(&v.PageReport, crawlId, page)
	case "inlinks":
		paginator.TotalPages = s.store.GetNumberOfPagesForInlinks(&v.PageReport, crawlId)
		v.InLinks = s.store.FindInLinks(v.PageReport.URL, crawlId, page)
	case "redirections":
		paginator.TotalPages = s.store.GetNumberOfPagesForRedirecting(&v.PageReport, crawlId)
		v.Redirects = s.store.FindPageReportsRedirectingToURL(v.PageReport.URL, crawlId, page)
	}

	if paginator.TotalPages == 0 {
		paginator.TotalPages = 1
	}

	if paginator.CurrentPage < paginator.TotalPages {
		paginator.NextPage = paginator.CurrentPage + 1
	}

	if paginator.CurrentPage > 1 {
		paginator.PreviousPage = paginator.CurrentPage - 1
	}

	v.Paginator = paginator

	return v
}

// Return channel of PageReports by error type
func (s *ReportService) GetPageReporsByIssueType(crawlId int64, eid string) <-chan *crawler.PageReport {
	if eid != "" {
		return s.store.FindAllPageReportsByCrawlIdAndErrorType(crawlId, eid)
	}

	return s.store.FindAllPageReportsByCrawlId(crawlId)
}

// Returns a channel of crawlable PageReports that can be included in a sitemap
func (s *ReportService) GetSitemapPageReports(crawlId int64) <-chan *crawler.PageReport {
	return s.store.FindSitemapPageReports(crawlId)
}
