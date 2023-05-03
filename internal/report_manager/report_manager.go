package report_manager

import (
	"sync"

	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporters"
	"github.com/stjudewashere/seonaut/internal/report_manager/sql_reporters"
)

type ReportManager struct {
	store              ReportManagerStore
	pageCallbacks      []*reporters.PageIssueReporter
	multipageCallbacks []sql_reporters.MultipageCallback
}

type ReportManagerStore interface {
	SaveIssues(<-chan *issue.Issue)
	PageReportsQuery(query string, args ...interface{}) <-chan *models.PageReport
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
func (rm *ReportManager) AddMultipageReporter(reporter sql_reporters.MultipageCallback) {
	rm.multipageCallbacks = append(rm.multipageCallbacks, reporter)
}

// Add an page issue reporter to the ReportManager.
// It will be used to create issues on each crawled page.
func (rm *ReportManager) AddPageReporter(reporter *reporters.PageIssueReporter) {
	rm.pageCallbacks = append(rm.pageCallbacks, reporter)
}

// CreateIssues uses the Reporters to create and save issues found in a crawl.
func (r *ReportManager) CreateIssues(crawl *models.Crawl) {
	issueCount := 0
	iStream := make(chan *issue.Issue)

	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		r.store.SaveIssues(iStream)
		wg.Done()
	}()

	for _, callback := range r.multipageCallbacks {
		reporter := callback(crawl)
		for p := range r.store.PageReportsQuery(reporter.Query, reporter.Parameters...) {
			iStream <- &issue.Issue{
				PageReportId: p.Id,
				CrawlId:      crawl.Id,
				ErrorType:    reporter.ErrorType,
			}

			issueCount++
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
