package report

import (
	"github.com/stjudewashere/seonaut/internal/pagereport"
)

type ReportStore interface {
	FindPageReportById(int) pagereport.PageReport
	FindErrorTypesByPage(int, int64) []string
	FindInLinks(string, int64, int) []pagereport.PageReport
	FindPageReportsRedirectingToURL(string, int64, int) []pagereport.PageReport
	FindAllPageReportsByCrawlIdAndErrorType(int64, string) <-chan *pagereport.PageReport
	FindAllPageReportsByCrawlId(int64) <-chan *pagereport.PageReport
	FindSitemapPageReports(int64) <-chan *pagereport.PageReport
	FindLinks(pageReport *pagereport.PageReport, cid int64, page int) []pagereport.Link
	FindExternalLinks(pageReport *pagereport.PageReport, cid int64, p int) []pagereport.Link

	GetNumberOfPagesForInlinks(*pagereport.PageReport, int64) int
	GetNumberOfPagesForRedirecting(*pagereport.PageReport, int64) int
	GetNumberOfPagesForLinks(*pagereport.PageReport, int64) int
	GetNumberOfPagesForExternalLinks(pageReport *pagereport.PageReport, cid int64) int
}

type ReportService struct {
	store ReportStore
}

type PageReportView struct {
	PageReport pagereport.PageReport
	ErrorTypes []string
	InLinks    []pagereport.PageReport
	Redirects  []pagereport.PageReport
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
func (s *ReportService) GetPageReporsByIssueType(crawlId int64, eid string) <-chan *pagereport.PageReport {
	if eid != "" {
		return s.store.FindAllPageReportsByCrawlIdAndErrorType(crawlId, eid)
	}

	return s.store.FindAllPageReportsByCrawlId(crawlId)
}

// Returns a channel of crawlable PageReports that can be included in a sitemap
func (s *ReportService) GetSitemapPageReports(crawlId int64) <-chan *pagereport.PageReport {
	return s.store.FindSitemapPageReports(crawlId)
}
