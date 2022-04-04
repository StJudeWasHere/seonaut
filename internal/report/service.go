package report

import (
	"github.com/stjudewashere/seonaut/internal/crawler"
)

type ReportStore interface {
	FindPageReportById(int) crawler.PageReport
	FindErrorTypesByPage(int, int64) []string
	FindInLinks(string, int64) []crawler.PageReport
	FindPageReportsRedirectingToURL(string, int64) []crawler.PageReport
	FindAllPageReportsByCrawlIdAndErrorType(int64, string) []crawler.PageReport
	FindAllPageReportsByCrawlId(int64) []crawler.PageReport
	FindSitemapPageReports(int64) []crawler.PageReport
}

type ReportService struct {
	store ReportStore
}

type PageReportView struct {
	PageReport crawler.PageReport
	ErrorTypes []string
	InLinks    []crawler.PageReport
	Redirects  []crawler.PageReport
}

func NewService(store ReportStore) *ReportService {
	return &ReportService{
		store: store,
	}
}

func (s *ReportService) GetPageReport(rid int, crawlId int64, tab string) *PageReportView {
	v := &PageReportView{
		PageReport: s.store.FindPageReportById(rid),
		ErrorTypes: s.store.FindErrorTypesByPage(rid, crawlId),
	}

	switch tab {
	case "inlinks":
		v.InLinks = s.store.FindInLinks(v.PageReport.URL, crawlId)
	case "redirections":
		v.Redirects = s.store.FindPageReportsRedirectingToURL(v.PageReport.URL, crawlId)
	}

	return v
}

func (s *ReportService) GetPageReporsByIssueType(crawlId int64, eid string) []crawler.PageReport {
	if eid != "" {
		return s.store.FindAllPageReportsByCrawlIdAndErrorType(crawlId, eid)
	}

	return s.store.FindAllPageReportsByCrawlId(crawlId)
}

func (s *ReportService) GetSitemapPageReports(crawlId int64) []crawler.PageReport {
	return s.store.FindSitemapPageReports(crawlId)
}
