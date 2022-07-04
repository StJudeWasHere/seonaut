package report_test

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

func (s *storage) FindAllPageReportsByCrawlIdAndErrorType(id int64, e string) <-chan *crawler.PageReport {
	prStream := make(chan *crawler.PageReport)
	go func() {
		defer close(prStream)
		prStream <- &crawler.PageReport{}
	}()

	return prStream
}

func (s *storage) FindAllPageReportsByCrawlId(id int64) <-chan *crawler.PageReport {
	prStream := make(chan *crawler.PageReport)

	go func() {
		defer close(prStream)
		prStream <- &crawler.PageReport{}
		prStream <- &crawler.PageReport{}
	}()

	return prStream
}

func (s *storage) FindSitemapPageReports(id int64) <-chan *crawler.PageReport {
	prStream := make(chan *crawler.PageReport)

	go func() {
		defer close(prStream)
		if id == crawlId {
			prStream <- &crawler.PageReport{}
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
