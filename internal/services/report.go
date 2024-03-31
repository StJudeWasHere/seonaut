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

		FindPageReportStyles(pageReport *models.PageReport, cid int64) []string
		FindPageReportScripts(pageReport *models.PageReport, cid int64) []string
		FindPageReportVideos(pageReport *models.PageReport, cid int64) []string
		FindPageReportAudios(pageReport *models.PageReport, cid int64) []string
		FindPageReportIframes(pageReport *models.PageReport, cid int64) []string
		FindPageReportImages(pageReport *models.PageReport, cid int64) []models.Image
		FindPageReportHreflangs(pageReport *models.PageReport, cid int64) []models.Hreflang

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
	v := &models.PageReportView{
		PageReport: s.store.FindPageReportById(rid),
		ErrorTypes: s.store.FindErrorTypesByPage(rid, crawlId),
	}

	v.PageReport.Hreflangs = s.store.FindPageReportHreflangs(&v.PageReport, crawlId)

	switch tab {
	case "internal":
		v.PageReport.InternalLinks = s.store.FindLinks(&v.PageReport, crawlId, page)
	case "external":
		v.PageReport.ExternalLinks = s.store.FindExternalLinks(&v.PageReport, crawlId, page)
	case "inlinks":
		v.InLinks = s.store.FindInLinks(v.PageReport.URL, crawlId, page)
	case "redirections":
		v.Redirects = s.store.FindPageReportsRedirectingToURL(v.PageReport.URL, crawlId, page)
	case "styles":
		v.PageReport.Styles = s.store.FindPageReportStyles(&v.PageReport, crawlId)
	case "scripts":
		v.PageReport.Scripts = s.store.FindPageReportScripts(&v.PageReport, crawlId)
	case "videos":
		v.PageReport.Videos = s.store.FindPageReportVideos(&v.PageReport, crawlId)
	case "audios":
		v.PageReport.Audios = s.store.FindPageReportAudios(&v.PageReport, crawlId)
	case "iframes":
		v.PageReport.Iframes = s.store.FindPageReportIframes(&v.PageReport, crawlId)
	case "images":
		v.PageReport.Images = s.store.FindPageReportImages(&v.PageReport, crawlId)
	}

	v.Paginator = s.getPaginator(&v.PageReport, crawlId, tab, page)

	return v
}

// Returns the paginator for the specific "tab".
func (s *ReportService) getPaginator(pageReport *models.PageReport, crawlId int64, tab string, page int) models.Paginator {
	paginator := models.Paginator{
		CurrentPage: page,
	}

	switch tab {
	case "internal":
		paginator.TotalPages = s.store.GetNumberOfPagesForLinks(pageReport, crawlId)
	case "external":
		paginator.TotalPages = s.store.GetNumberOfPagesForExternalLinks(pageReport, crawlId)
	case "inlinks":
		paginator.TotalPages = s.store.GetNumberOfPagesForInlinks(pageReport, crawlId)
	case "redirections":
		paginator.TotalPages = s.store.GetNumberOfPagesForRedirecting(pageReport, crawlId)
	default:
		paginator.TotalPages = 1
	}

	if paginator.CurrentPage < paginator.TotalPages {
		paginator.NextPage = paginator.CurrentPage + 1
	}

	if paginator.CurrentPage > 1 {
		paginator.PreviousPage = paginator.CurrentPage - 1
	}

	return paginator
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
