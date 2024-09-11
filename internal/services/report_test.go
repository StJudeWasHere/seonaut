package services_test

import (
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/services"
)

const (
	crawlId         = 1
	reportId        = 10
	errorType       = "ERROR"
	tabInlinks      = "inlinks"
	tabRedirections = "redirections"
	page            = 1
)

type reportTestRepository struct{}

func (s *reportTestRepository) CountByMediaType(i int64) *models.CountList {
	return &models.CountList{}
}

func (s *reportTestRepository) CountByStatusCode(i int64) *models.CountList {
	return &models.CountList{}
}

func (s *reportTestRepository) CountByCanonical(i int64) int {
	return 0
}

func (s *reportTestRepository) CountImagesAlt(i int64) *models.AltCount {
	return &models.AltCount{}
}

func (s *reportTestRepository) CountScheme(i int64) *models.SchemeCount {
	return &models.SchemeCount{}
}

func (s *reportTestRepository) CountByNonCanonical(i int64) int {
	return 0
}

func (s *reportTestRepository) FindExternalLinks(pageReport *models.PageReport, cid int64, p int) []models.Link {
	return []models.Link{}
}

func (s *reportTestRepository) GetNumberOfPagesForExternalLinks(pageReport *models.PageReport, cid int64) int {
	return 0
}

func (s *reportTestRepository) GetNumberOfPagesForLinks(pageReport *models.PageReport, cid int64) int {
	return 0
}

func (s *reportTestRepository) GetNumberOfPagesForRedirecting(pageReport *models.PageReport, cid int64) int {
	return 0
}

func (s *reportTestRepository) FindLinks(pageReport *models.PageReport, cid int64, p int) []models.InternalLink {
	return []models.InternalLink{}
}

func (s *reportTestRepository) FindPageReportById(id int) models.PageReport {
	return models.PageReport{Id: reportId}
}

func (s *reportTestRepository) FindErrorTypesByPage(reportId int, crawlId int64) []string {
	return []string{errorType}
}

func (s *reportTestRepository) FindInLinks(u string, id int64, page int) []models.InternalLink {
	return []models.InternalLink{{PageReport: models.PageReport{Id: reportId}}}
}

func (s *reportTestRepository) GetNumberOfPagesForInlinks(pageReport *models.PageReport, cid int64) int {
	return 1
}

func (s *reportTestRepository) FindPageReportsRedirectingToURL(u string, id int64, page int) []models.PageReport {
	return []models.PageReport{{Id: reportId}}
}

func (s *reportTestRepository) FindAllPageReportsByCrawlIdAndErrorType(id int64, e string) <-chan *models.PageReport {
	prStream := make(chan *models.PageReport)
	go func() {
		defer close(prStream)
		prStream <- &models.PageReport{}
	}()

	return prStream
}

func (s *reportTestRepository) FindAllPageReportsByCrawlId(id int64) <-chan *models.PageReport {
	prStream := make(chan *models.PageReport)

	go func() {
		defer close(prStream)
		prStream <- &models.PageReport{}
		prStream <- &models.PageReport{}
	}()

	return prStream
}

func (s *reportTestRepository) FindSitemapPageReports(id int64) <-chan *models.PageReport {
	prStream := make(chan *models.PageReport)

	go func() {
		defer close(prStream)
		if id == crawlId {
			prStream <- &models.PageReport{}
		}
	}()

	return prStream
}

func (s *reportTestRepository) FindPaginatedPageReports(cid int64, p int, term string) []models.PageReport {
	return []models.PageReport{}
}

func (s *reportTestRepository) GetNumberOfPagesForPageReport(cid int64, term string) int {
	return 0
}

func (s *reportTestRepository) GetStatusCodeByDepth(crawlId int64) []models.StatusCodeByDepth {
	return []models.StatusCodeByDepth{}
}

func (s *reportTestRepository) FindPageReportStyles(pageReport *models.PageReport, cid int64) []string {
	return []string{}
}
func (s *reportTestRepository) FindPageReportScripts(pageReport *models.PageReport, cid int64) []string {
	return []string{}
}
func (s *reportTestRepository) FindPageReportVideos(pageReport *models.PageReport, cid int64) []models.Video {
	return []models.Video{}
}
func (s *reportTestRepository) FindPageReportAudios(pageReport *models.PageReport, cid int64) []string {
	return []string{}
}
func (s *reportTestRepository) FindPageReportIframes(pageReport *models.PageReport, cid int64) []string {
	return []string{}
}
func (s *reportTestRepository) FindPageReportImages(pageReport *models.PageReport, cid int64) []models.Image {
	return []models.Image{}
}
func (s *reportTestRepository) FindPageReportHreflangs(pageReport *models.PageReport, cid int64) []models.Hreflang {
	return []models.Hreflang{}
}

var reportservice = services.NewReportService(&reportTestRepository{})

func TestGetSitemapPageReports(t *testing.T) {
	prStream := reportservice.GetSitemapPageReports(crawlId)
	l := 0

	for range prStream {
		l++
	}

	if l != 1 {
		t.Errorf("GetSitemapPageReports: %d != 1", l)
	}
}

func TestGetPageReporsByIssueType(t *testing.T) {
	prStream := reportservice.GetPageReporsByIssueType(crawlId, "")
	l := 0
	for range prStream {
		l++
	}

	if l != 2 {
		t.Errorf("GetPageReporsByIssueType: %d != 2", l)
	}

	pe := reportservice.GetPageReporsByIssueType(crawlId, errorType)
	l = 0
	for range pe {
		l++
	}

	if l != 1 {
		t.Errorf("GetPageReporsByIssueType: %d != 1", l)
	}
}

func TestGetPageReport(t *testing.T) {
	v := reportservice.GetPageReport(reportId, crawlId, tabInlinks, page)
	if v.PageReport.Id != reportId {
		t.Errorf("GetPageReport: %d != %d", v.PageReport.Id, reportId)
	}

	if len(v.ErrorTypes) != 1 {
		t.Errorf("v.ErrorTypes: %d != 1", len(v.ErrorTypes))
	}

	if len(v.InLinks) != 1 {
		t.Errorf("v.InLinks: %d != 1", len(v.InLinks))
	}

	if len(v.Redirects) != 0 {
		t.Errorf("v.Redirects: %d != 0", len(v.Redirects))
	}

	vr := reportservice.GetPageReport(reportId, crawlId, tabRedirections, page)
	if len(vr.InLinks) != 0 {
		t.Errorf("v.InLinks: %d != 0", len(vr.InLinks))
	}

	if len(vr.Redirects) != 1 {
		t.Errorf("v.Redirects: %d != 1", len(vr.Redirects))
	}
}
