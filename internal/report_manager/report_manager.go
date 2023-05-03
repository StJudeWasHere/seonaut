package report_manager

import (
	"sync"

	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporters"
)

type MultipageCallback func(c *models.Crawl) *MultipageIssueReporter

type MultipageIssueReporter struct {
	Pstream   <-chan *models.PageReport
	ErrorType int
}

type ReportManager struct {
	store              ReportManagerStore
	pageCallbacks      []*reporters.PageIssueReporter
	multipageCallbacks []MultipageCallback
}

type ReportManagerStore interface {
	SaveIssues(<-chan *issue.Issue)
}

// Create a new ReportManager with no issue reporters.
func NewReportManager(s ReportManagerStore) *ReportManager {
	return &ReportManager{
		store: s,
	}
}

// Add a multi-page issue reporter to the ReportManager. Multi-page reporters are used to detect
// issues that affect multiple pages. It will be used when creating the multi page issues once all
// the pages have been crawled.
func (rm *ReportManager) AddMultipageReporter(reporter MultipageCallback) {
	rm.multipageCallbacks = append(rm.multipageCallbacks, reporter)
}

// Add an page issue reporter to the ReportManager.
// It will be used to create issues on each crawled page.
func (rm *ReportManager) AddPageReporter(reporter *reporters.PageIssueReporter) {
	rm.pageCallbacks = append(rm.pageCallbacks, reporter)
}

// CreateIssues uses the Reporters to create and save issues found in a crawl.
func (r *ReportManager) CreateIssues(crawl *models.Crawl) {
	iStream := make(chan *issue.Issue)
	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		r.store.SaveIssues(iStream)
		wg.Done()
	}()

	for _, callback := range r.multipageCallbacks {
		reporter := callback(crawl)
		for p := range reporter.Pstream {
			iStream <- &issue.Issue{
				PageReportId: p.Id,
				CrawlId:      crawl.Id,
				ErrorType:    reporter.ErrorType,
			}
		}
	}

	close(iStream)

	wg.Wait()
}

// CreatePageIssues loops the page reporters calling the callback function
// and creating the issues found in the PageReport.
func (r *ReportManager) CreatePageIssues(p *models.PageReport, crawl *models.Crawl) {
	iStream := make(chan *issue.Issue)
	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		r.store.SaveIssues(iStream)
		wg.Done()
	}()

	for _, c := range r.pageCallbacks {
		if c.Callback(p) == true {
			iStream <- &issue.Issue{
				PageReportId: p.Id,
				CrawlId:      crawl.Id,
				ErrorType:    c.ErrorType,
			}
		}
	}

	close(iStream)

	wg.Wait()
}
