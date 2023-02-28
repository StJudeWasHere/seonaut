package report_test

import (
	"testing"

	"github.com/stjudewashere/seonaut/internal/pagereport"
	"github.com/stjudewashere/seonaut/internal/report"
)

const (
	crawlId         = 1
	reportId        = 10
	errorType       = "ERROR"
	tabInlinks      = "inlinks"
	tabRedirections = "redirections"
	page            = 1
)

type storage struct{}

func (s *storage) CountByMediaType(i int64) report.CountList {
	return report.CountList{}
}

func (s *storage) CountByStatusCode(i int64) report.CountList {
	return report.CountList{}
}

func (s *storage) CountByCanonical(i int64) int {
	return 0
}

func (s *storage) CountImagesAlt(i int64) *report.AltCount {
	return &report.AltCount{}
}

func (s *storage) CountScheme(i int64) *report.SchemeCount {
	return &report.SchemeCount{}
}

func (s *storage) CountByNonCanonical(i int64) int {
	return 0
}

func (s *storage) FindExternalLinks(pageReport *pagereport.PageReport, cid int64, p int) []pagereport.Link {
	return []pagereport.Link{}
}

func (s *storage) GetNumberOfPagesForExternalLinks(pageReport *pagereport.PageReport, cid int64) int {
	return 0
}

func (s *storage) GetNumberOfPagesForLinks(pageReport *pagereport.PageReport, cid int64) int {
	return 0
}

func (s *storage) GetNumberOfPagesForRedirecting(pageReport *pagereport.PageReport, cid int64) int {
	return 0
}

func (s *storage) FindLinks(pageReport *pagereport.PageReport, cid int64, p int) []pagereport.InternalLink {
	return []pagereport.InternalLink{}
}

func (s *storage) FindPageReportById(id int) pagereport.PageReport {
	return pagereport.PageReport{Id: reportId}
}

func (s *storage) FindErrorTypesByPage(reportId int, crawlId int64) []string {
	return []string{errorType}
}

func (s *storage) FindInLinks(u string, id int64, page int) []pagereport.InternalLink {
	return []pagereport.InternalLink{{PageReport: pagereport.PageReport{Id: reportId}}}
}

func (s *storage) GetNumberOfPagesForInlinks(pageReport *pagereport.PageReport, cid int64) int {
	return 1
}

func (s *storage) FindPageReportsRedirectingToURL(u string, id int64, page int) []pagereport.PageReport {
	return []pagereport.PageReport{{Id: reportId}}
}

func (s *storage) FindAllPageReportsByCrawlIdAndErrorType(id int64, e string) <-chan *pagereport.PageReport {
	prStream := make(chan *pagereport.PageReport)
	go func() {
		defer close(prStream)
		prStream <- &pagereport.PageReport{}
	}()

	return prStream
}

func (s *storage) FindAllPageReportsByCrawlId(id int64) <-chan *pagereport.PageReport {
	prStream := make(chan *pagereport.PageReport)

	go func() {
		defer close(prStream)
		prStream <- &pagereport.PageReport{}
		prStream <- &pagereport.PageReport{}
	}()

	return prStream
}

func (s *storage) FindSitemapPageReports(id int64) <-chan *pagereport.PageReport {
	prStream := make(chan *pagereport.PageReport)

	go func() {
		defer close(prStream)
		if id == crawlId {
			prStream <- &pagereport.PageReport{}
		}
	}()

	return prStream
}

var service = report.NewService(&storage{})

func TestGetSitemapPageReports(t *testing.T) {
	prStream := service.GetSitemapPageReports(crawlId)
	l := 0

	for range prStream {
		l++
	}

	if l != 1 {
		t.Errorf("GetSitemapPageReports: %d != 1", l)
	}
}

func TestGetPageReporsByIssueType(t *testing.T) {
	prStream := service.GetPageReporsByIssueType(crawlId, "")
	l := 0
	for range prStream {
		l++
	}

	if l != 2 {
		t.Errorf("GetPageReporsByIssueType: %d != 2", l)
	}

	pe := service.GetPageReporsByIssueType(crawlId, errorType)
	l = 0
	for range pe {
		l++
	}

	if l != 1 {
		t.Errorf("GetPageReporsByIssueType: %d != 1", l)
	}
}

func TestGetPageReport(t *testing.T) {
	v := service.GetPageReport(reportId, crawlId, tabInlinks, page)
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

	vr := service.GetPageReport(reportId, crawlId, tabRedirections, page)
	if len(vr.InLinks) != 0 {
		t.Errorf("v.InLinks: %d != 0", len(vr.InLinks))
	}

	if len(vr.Redirects) != 1 {
		t.Errorf("v.Redirects: %d != 1", len(vr.Redirects))
	}
}
