// The report_manager takes care of running the issue reporters against the crawled pages.
// There are two different types of issue reporters. On one hand there's the PageIssueReporters,
// which are run against single pages as they are crawled. This checks can detect issues in the
// headers and body of the PageReport, such as wrong headers or missing tags.
// On the other hand there is the MultipageIssuReporters, which can run checks that affect multiple
// pages, such as duplicated titles.
package services

import (
	"net/http"
	"sync"

	"golang.org/x/net/html"

	"github.com/stjudewashere/seonaut/internal/models"
)

type (
	ReportManagerRepository interface {
		SaveIssues(<-chan *models.Issue)
	}

	ReportManager struct {
		repository         ReportManagerRepository
		pageCallbacks      []*models.PageIssueReporter
		multipageCallbacks []models.MultipageCallback
	}
)

// Create a new ReportManager with no issue reporters.
func NewReportManager(r ReportManagerRepository) *ReportManager {
	return &ReportManager{
		repository: r,
	}
}

// Add an page issue reporter to the ReportManager.
// It will be used to create issues on each crawled page.
func (rm *ReportManager) AddPageReporter(reporter *models.PageIssueReporter) {
	rm.pageCallbacks = append(rm.pageCallbacks, reporter)
}

// Add a multi-page issue reporter to the ReportManager. Multi-page reporters are used to detect
// issues that affect multiple pages. It will be used when creating the multi page issues once all
// the pages have been crawled.
func (rm *ReportManager) AddMultipageReporter(reporter models.MultipageCallback) {
	rm.multipageCallbacks = append(rm.multipageCallbacks, reporter)
}

// CreatePageIssues loops the page reporters calling the callback function
// and creating the issues found in the PageReport.
func (r *ReportManager) CreatePageIssues(p *models.PageReport, htmlNode *html.Node, header *http.Header, crawl *models.Crawl) {
	iStream := make(chan *models.Issue)
	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		r.repository.SaveIssues(iStream)
		wg.Done()
	}()

	for _, c := range r.pageCallbacks {
		if c.Callback(p, htmlNode, header) {
			iStream <- &models.Issue{
				PageReportId: p.Id,
				CrawlId:      crawl.Id,
				ErrorType:    c.ErrorType,
			}
		}
	}

	close(iStream)

	wg.Wait()
}

// CreateMultipageIssues uses the Reporters to create and save issues found in a crawl.
func (r *ReportManager) CreateMultipageIssues(crawl *models.Crawl) {
	iStream := make(chan *models.Issue)
	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		r.repository.SaveIssues(iStream)
		wg.Done()
	}()

	for _, callback := range r.multipageCallbacks {
		reporter := callback(crawl)
		for pid := range reporter.Pstream {
			iStream <- &models.Issue{
				PageReportId: pid,
				CrawlId:      crawl.Id,
				ErrorType:    reporter.ErrorType,
			}
		}
	}

	close(iStream)

	wg.Wait()
}
