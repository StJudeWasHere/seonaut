package services

import (
	"errors"

	"github.com/stjudewashere/seonaut/internal/models"
)

type (
	ReportServiceStorage interface {
		FindPageReportById(int) models.PageReport
		FindErrorTypesByPage(int, int64) []string
		FindInLinks(string, int64, int) []models.InternalLink
		FindPageReportsRedirectingToURL(string, int64, int) []models.PageReport
		FindAllPageReportsByCrawlIdAndErrorType(int64, string) <-chan *models.PageReport
		FindAllPageReportsByCrawlId(int64) <-chan *models.PageReport
		FindSitemapPageReports(int64) <-chan *models.PageReport
		FindLinks(pageReport *models.PageReport, cid int64, page int) []models.InternalLink
		FindExternalLinks(pageReport *models.PageReport, cid int64, p int) []models.Link
		FindPaginatedPageReports(cid int64, p int, term string) []models.PageReport

		GetNumberOfPagesForPageReport(cid int64, term string) int
		GetNumberOfPagesForInlinks(*models.PageReport, int64) int
		GetNumberOfPagesForRedirecting(*models.PageReport, int64) int
		GetNumberOfPagesForLinks(*models.PageReport, int64) int
		GetNumberOfPagesForExternalLinks(pageReport *models.PageReport, cid int64) int
	}

	ReportService struct {
		store ReportServiceStorage
	}
)

func NewReportService(store ReportServiceStorage) *ReportService {
	return &ReportService{store: store}
}

// Returns a PageReportView by PageReport Id and Crawl Id.
// It also loads the data specified in the tab paramater.
func (s *ReportService) GetPageReport(rid int, crawlId int64, tab string, page int) *models.PageReportView {
	paginator := models.Paginator{
		CurrentPage: page,
	}

	v := &models.PageReportView{
		PageReport: s.store.FindPageReportById(rid),
		ErrorTypes: s.store.FindErrorTypesByPage(rid, crawlId),
	}

	switch tab {
	case "internal":
		paginator.TotalPages = s.store.GetNumberOfPagesForLinks(&v.PageReport, crawlId)
		v.PageReport.InternalLinks = s.store.FindLinks(&v.PageReport, crawlId, page)
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

// Return channel of PageReports by error type.
func (s *ReportService) GetPageReporsByIssueType(crawlId int64, eid string) <-chan *models.PageReport {
	if eid != "" {
		return s.store.FindAllPageReportsByCrawlIdAndErrorType(crawlId, eid)
	}

	return s.store.FindAllPageReportsByCrawlId(crawlId)
}

// Returns a PaginatorView with the corresponding page reports.
func (s *ReportService) GetPaginatedReports(crawlId int64, currentPage int, term string) (models.PaginatorView, error) {
	paginator := models.Paginator{
		TotalPages:  s.store.GetNumberOfPagesForPageReport(crawlId, term),
		CurrentPage: currentPage,
	}

	if currentPage < 1 || (paginator.TotalPages > 0 && currentPage > paginator.TotalPages) {
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
		PageReports: s.store.FindPaginatedPageReports(crawlId, currentPage, term),
	}

	return paginatorView, nil
}

// Returns a channel of crawlable PageReports that can be included in a sitemap.
func (s *ReportService) GetSitemapPageReports(crawlId int64) <-chan *models.PageReport {
	return s.store.FindSitemapPageReports(crawlId)
}
