package report_manager_test

import (
	"testing"

	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
)

const (
	crawlId      = 1
	pageReportId = 1
	errorType    = 1
)

// Mock storage contains an Issues slice so we can test if issues are being received.
type mockStorage struct {
	Issues []*issue.Issue
}

// SaveIssues appends the issue to the Issues slice.
func (s *mockStorage) SaveIssues(c <-chan *issue.Issue) {
	for i := range c {
		s.Issues = append(s.Issues, i)
	}
}

// Add a PageReporter and test if new issue is sent to the storage.
func TestCreatePageIssuesCreatesIssue(t *testing.T) {

	// Create a new mockStorage and report_manager service.
	storage := &mockStorage{}
	service := report_manager.NewReportManager(storage)

	// Add a new PageReporter that detects an issue.
	service.AddPageReporter(
		&report_manager.PageIssueReporter{
			ErrorType: errorType,
			Callback: func(pageReport *models.PageReport) bool {
				return true
			},
		})

	// Create a PageReport and Crawl.
	pageReport := &models.PageReport{Id: pageReportId}
	crawl := &models.Crawl{Id: crawlId}

	// Create the PageIssues should run the PageIssueReporter that returns true
	// indicating an issue was found, so a new issue should be created and added
	// to the mockStorage.
	service.CreatePageIssues(pageReport, crawl)

	// The storage should contain exactly one issue.
	if len(storage.Issues) != 1 {
		t.Errorf("CreatePageIsssues: %d != 1", len(storage.Issues))
	}

	issue := storage.Issues[0]

	// Make sure the created issue contains the right ids.
	if issue.PageReportId != pageReportId {
		t.Errorf("CreatePageIsssues: PageReportId %d != %d", issue.PageReportId, pageReportId)
	}

	if issue.CrawlId != crawlId {
		t.Errorf("CreatePageIsssues: crawlId %d != %d", issue.CrawlId, crawlId)
	}

	if issue.ErrorType != errorType {
		t.Errorf("CreatePageIsssues: crawlId %d != %d", issue.ErrorType, errorType)
	}
}

// Add a PageReporter and test if new issue is not sent to the storage.
func TestCreatePageIssuesDoesNotCreateIssue(t *testing.T) {

	// Create a new mockStorage and report_manager service.
	storage := &mockStorage{}
	service := report_manager.NewReportManager(storage)

	// Add a new PageReporter that detects an issue.
	service.AddPageReporter(
		&report_manager.PageIssueReporter{
			ErrorType: errorType,
			Callback: func(pageReport *models.PageReport) bool {
				return false
			},
		})

	// Create a PageReport and Crawl.
	pageReport := &models.PageReport{Id: pageReportId}
	crawl := &models.Crawl{Id: crawlId}

	// Create the PageIssues should run the PageIssueReporter that returns false
	// indicating an issue was not found and will not be created.
	service.CreatePageIssues(pageReport, crawl)

	// The storage issues slice should be empty.
	if len(storage.Issues) != 0 {
		t.Errorf("CreatePageIsssues: DoesNotCreateIssue: %d != 0", len(storage.Issues))
	}
}

// Add a MultipageReporter and test if an issue is created.
func TestCreateMultiPageIssues(t *testing.T) {
	// Create a new mockStorage and report_manager service.
	storage := &mockStorage{}
	service := report_manager.NewReportManager(storage)

	service.AddMultipageReporter(
		func(c *models.Crawl) *report_manager.MultipageIssueReporter {
			stream := make(chan int64)

			go func() {
				stream <- pageReportId
				close(stream)
			}()

			return &report_manager.MultipageIssueReporter{
				Pstream:   stream,
				ErrorType: errorType,
			}
		},
	)

	crawl := &models.Crawl{Id: crawlId}

	service.CreateMultipageIssues(crawl)

	// The storage should contain exactly one issue.
	if len(storage.Issues) != 1 {
		t.Errorf("CreatePageIsssues: %d != 1", len(storage.Issues))
	}

	issue := storage.Issues[0]

	// Make sure the created issue contains the right ids.
	if issue.PageReportId != pageReportId {
		t.Errorf("CreatePageIsssues: PageReportId %d != %d", issue.PageReportId, pageReportId)
	}

	if issue.CrawlId != crawlId {
		t.Errorf("CreatePageIsssues: crawlId %d != %d", issue.CrawlId, crawlId)
	}

	if issue.ErrorType != errorType {
		t.Errorf("CreatePageIsssues: crawlId %d != %d", issue.ErrorType, errorType)
	}
}
