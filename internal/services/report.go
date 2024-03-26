package services

import (
	"errors"
	"fmt"
	"log"

	"github.com/stjudewashere/seonaut/internal/models"
)

type ReportServiceCache interface {
	Set(key string, v interface{}) error
	Get(key string, v interface{}) error
	Delete(key string) error
}

type ReportStore interface {
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

	CountByMediaType(int64) *models.CountList
	CountByStatusCode(int64) *models.CountList

	CountByCanonical(int64) int
	CountImagesAlt(int64) *models.AltCount
	CountScheme(int64) *models.SchemeCount
	CountByNonCanonical(int64) int
	GetStatusCodeByDepth(crawlId int64) []models.StatusCodeByDepth
}

type ReportService struct {
	store ReportStore
	cache ReportServiceCache
}

func NewReportService(store ReportStore, cache ReportServiceCache) *ReportService {
	return &ReportService{
		store: store,
		cache: cache,
	}
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

// Returns a CountList with the PageReport's media type count.
func (s *ReportService) GetMediaCount(crawlId int64) *models.CountList {
	key := fmt.Sprintf("media-%d", crawlId)
	v := &models.CountList{}
	if err := s.cache.Get(key, v); err != nil {
		v = s.store.CountByMediaType(crawlId)
		if err := s.cache.Set(key, v); err != nil {
			log.Printf("GetMediaCount: cacheSet: %v\n", err)
		}
	}

	return v
}

// Returns a CountList with the PageReport's status code count.
func (s *ReportService) GetStatusCount(crawlId int64) *models.CountList {
	key := fmt.Sprintf("status-%d", crawlId)
	v := &models.CountList{}
	if err := s.cache.Get(key, v); err != nil {
		v = s.store.CountByStatusCode(crawlId)
		if err := s.cache.Set(key, v); err != nil {
			log.Printf("GetStatusCount: cacheSet: %v\n", err)
		}
	}

	return v
}

// Returns the count Images with and without the alt attribute.
func (s *ReportService) GetImageAltCount(crawlId int64) *models.AltCount {
	key := fmt.Sprintf("alt-%d", crawlId)
	v := &models.AltCount{}
	if err := s.cache.Get(key, v); err != nil {
		v = s.store.CountImagesAlt(crawlId)
		if err := s.cache.Set(key, v); err != nil {
			log.Printf("GetImageAltCount: cacheSet: %v\n", err)
		}
	}

	return v
}

// Returns the count of PageReports with and without https.
func (s *ReportService) GetSchemeCount(crawlId int64) *models.SchemeCount {
	key := fmt.Sprintf("scheme-%d", crawlId)
	v := &models.SchemeCount{}
	if err := s.cache.Get(key, v); err != nil {
		v = s.store.CountScheme(crawlId)
		if err := s.cache.Set(key, v); err != nil {
			log.Printf("GetSchemeCount: cacheSet: %v\n", err)
		}
	}
	return v
}

// Returns a count of PageReports that are canonical or not.
func (s *ReportService) GetCanonicalCount(crawlId int64) *models.CanonicalCount {
	key := fmt.Sprintf("canonical-%d", crawlId)
	c := &models.CanonicalCount{}
	if err := s.cache.Get(key, c); err != nil {
		c.Canonical = s.store.CountByCanonical(crawlId)
		c.NonCanonical = s.store.CountByNonCanonical(crawlId)
		if err := s.cache.Set(key, c); err != nil {
			log.Printf("GetCanonicalCount: cacheSet: %v\n", err)
		}
	}

	return c
}

func (s *ReportService) BuildCrawlCache(crawl *models.Crawl) {
	media := s.store.CountByMediaType(crawl.Id)
	if err := s.cache.Set(fmt.Sprintf("media-%d", crawl.Id), media); err != nil {
		log.Printf("BuildDashboardCache: Media: %v\n", err)
	}

	status := s.store.CountByStatusCode(crawl.Id)
	if err := s.cache.Set(fmt.Sprintf("status-%d", crawl.Id), status); err != nil {
		log.Printf("BuildDashboardCache: Status: %v\n", err)
	}

	alt := s.store.CountImagesAlt(crawl.Id)
	if err := s.cache.Set(fmt.Sprintf("alt-%d", crawl.Id), alt); err != nil {
		log.Printf("BuildDashboardCache: Alt: %v\n", err)
	}

	scheme := s.store.CountScheme(crawl.Id)
	if err := s.cache.Set(fmt.Sprintf("scheme-%d", crawl.Id), scheme); err != nil {
		log.Printf("BuildDashboardCache: Scheme: %v\n", err)
	}

	canonical := &models.CanonicalCount{
		Canonical:    s.store.CountByCanonical(crawl.Id),
		NonCanonical: s.store.CountByNonCanonical(crawl.Id),
	}

	if err := s.cache.Set(fmt.Sprintf("canonical-%d", crawl.Id), canonical); err != nil {
		log.Printf("BuildDashboardCache: Canonical: %v\n", err)
	}
}

func (s *ReportService) RemoveCrawlCache(crawl *models.Crawl) {
	if err := s.cache.Delete(fmt.Sprintf("media-%d", crawl.Id)); err != nil {
		log.Printf("DeleteDashboardCache: Media: %v\n", err)
	}

	if err := s.cache.Delete(fmt.Sprintf("status-%d", crawl.Id)); err != nil {
		log.Printf("DeleteDashboardCache: Status: %v\n", err)
	}

	if err := s.cache.Delete(fmt.Sprintf("alt-%d", crawl.Id)); err != nil {
		log.Printf("DeleteDashboardCache: Alt: %v\n", err)
	}

	if err := s.cache.Delete(fmt.Sprintf("scheme-%d", crawl.Id)); err != nil {
		log.Printf("DeleteDashboardCache: Scheme: %v\n", err)
	}

	if err := s.cache.Delete(fmt.Sprintf("canonical-%d", crawl.Id)); err != nil {
		log.Printf("DeleteDashboardCache: Canonical: %v\n", err)
	}
}

func (s *ReportService) GetStatusCodeByDepth(crawlId int64) []models.StatusCodeByDepth {
	key := fmt.Sprintf("status-by-depth-%d", crawlId)
	v := []models.StatusCodeByDepth{}
	if err := s.cache.Get(key, &v); err != nil {
		v = s.store.GetStatusCodeByDepth(crawlId)
		if err := s.cache.Set(key, &v); err != nil {
			log.Printf("GetStatusCodeByDepth: cacheSet: %v\n", err)
		}
	}

	return v
}
