package user

import (
	"testing"

	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/report"
)

const (
	crawlId         = 1
	reportId        = 10
	errorType       = "ERROR"
	tabInlinks      = "inlinks"
	tabRedirections = "redirections"
)

type storage struct{}

func (s *storage) FindPageReportById(id int) crawler.PageReport {
	return crawler.PageReport{Id: reportId}
}

func (s *storage) FindErrorTypesByPage(reportId int, crawlId int64) []string {
	return []string{errorType}
}

func (s *storage) FindInLinks(u string, id int64) []crawler.PageReport {
	return []crawler.PageReport{crawler.PageReport{Id: reportId}}
}

func (s *storage) FindPageReportsRedirectingToURL(u string, id int64) []crawler.PageReport {
	return []crawler.PageReport{crawler.PageReport{Id: reportId}}
}

func (s *storage) FindAllPageReportsByCrawlIdAndErrorType(id int64, e string) []crawler.PageReport {
	return []crawler.PageReport{crawler.PageReport{}}
}

func (s *storage) FindAllPageReportsByCrawlId(id int64) []crawler.PageReport {
	return []crawler.PageReport{
		crawler.PageReport{},
		crawler.PageReport{},
	}
}

func (s *storage) FindSitemapPageReports(id int64) []crawler.PageReport {
	r := []crawler.PageReport{}

	if id == crawlId {
		r = append(r, crawler.PageReport{})
		return r
	}

	return r
}

var service = report.NewService(&storage{})

func TestGetSitemapPageReports(t *testing.T) {
	p := service.GetSitemapPageReports(crawlId)
	if len(p) != 1 {
		t.Errorf("GetSitemapPageReports: %d != 1", len(p))
	}
}

func TestGetPageReporsByIssueType(t *testing.T) {
	p := service.GetPageReporsByIssueType(crawlId, "")
	if len(p) != 2 {
		t.Errorf("GetPageReporsByIssueType: %d != 2", len(p))
	}

	pe := service.GetPageReporsByIssueType(crawlId, errorType)
	if len(pe) != 1 {
		t.Errorf("GetPageReporsByIssueType: %d != 1", len(pe))
	}
}

func TestGetPageReport(t *testing.T) {
	v := service.GetPageReport(reportId, crawlId, tabInlinks)
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

	vr := service.GetPageReport(reportId, crawlId, tabRedirections)
	if len(vr.InLinks) != 0 {
		t.Errorf("v.InLinks: %d != 0", len(vr.InLinks))
	}

	if len(vr.Redirects) != 1 {
		t.Errorf("v.Redirects: %d != 1", len(vr.Redirects))
	}
}
