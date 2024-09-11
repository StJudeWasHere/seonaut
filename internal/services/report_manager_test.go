package services_test

import (
	"net/http"
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/services"

	"golang.org/x/net/html"
)

const (
	reporterCrawlId   = 1
	pageReportId      = 1
	reporterErrorType = 1
)

// Mock repository contains an Issues slice so we can test if issues are being received.
type reportManagerTestRepository struct {
	Issues []*models.Issue
}

// SaveIssues appends the issue to the Issues slice.
func (s *reportManagerTestRepository) SaveIssues(c <-chan *models.Issue) {
	for i := range c {
		s.Issues = append(s.Issues, i)
	}
}

// Add a PageReporter and test if new issue is sent to the repository.
func TestCreatePageIssuesCreatesIssue(t *testing.T) {

	// Create a new reportManagerTestRepository and report_manager service.
	repository := &reportManagerTestRepository{}
	service := services.NewReportManager(repository)

	// Add a new PageReporter that detects an issue.
	service.AddPageReporter(
		&models.PageIssueReporter{
			ErrorType: reporterErrorType,
			Callback: func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
				return true
			},
		})

	// Create a PageReport and Crawl.
	pageReport := &models.PageReport{Id: pageReportId}
	crawl := &models.Crawl{Id: reporterCrawlId}

	// Create the PageIssues should run the PageIssueReporter that returns true
	// indicating an issue was found, so a new issue should be created and added
	// to the reportManagerTestRepository.
	service.CreatePageIssues(pageReport, &html.Node{}, &http.Header{}, crawl)

	// The repository should contain exactly one issue.
	if len(repository.Issues) != 1 {
		t.Errorf("CreatePageIsssues: %d != 1", len(repository.Issues))
	}

	issue := repository.Issues[0]

	// Make sure the created issue contains the right ids.
	if issue.PageReportId != pageReportId {
		t.Errorf("CreatePageIsssues: PageReportId %d != %d", issue.PageReportId, pageReportId)
	}

	if issue.CrawlId != reporterCrawlId {
		t.Errorf("CreatePageIsssues: reporterCrawlId %d != %d", issue.CrawlId, reporterCrawlId)
	}

	if issue.ErrorType != reporterErrorType {
		t.Errorf("CreatePageIsssues: reporterCrawlId %d != %d", issue.ErrorType, reporterErrorType)
	}
}

// Add a PageReporter and test if new issue is not sent to the repository.
func TestCreatePageIssuesDoesNotCreateIssue(t *testing.T) {

	// Create a new reportManagerTestRepository and report_manager service.
	repository := &reportManagerTestRepository{}
	service := services.NewReportManager(repository)

	// Add a new PageReporter that detects an issue.
	service.AddPageReporter(
		&models.PageIssueReporter{
			ErrorType: reporterErrorType,
			Callback: func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
				return false
			},
		})

	// Create a PageReport and Crawl.
	pageReport := &models.PageReport{Id: pageReportId}
	crawl := &models.Crawl{Id: reporterCrawlId}

	// Create the PageIssues should run the PageIssueReporter that returns false
	// indicating an issue was not found and will not be created.
	service.CreatePageIssues(pageReport, &html.Node{}, &http.Header{}, crawl)

	// The repository issues slice should be empty.
	if len(repository.Issues) != 0 {
		t.Errorf("CreatePageIsssues: DoesNotCreateIssue: %d != 0", len(repository.Issues))
	}
}

// Add a MultipageReporter and test if an issue is created.
func TestCreateMultiPageIssues(t *testing.T) {
	// Create a new reportManagerTestRepository and report_manager service.
	repository := &reportManagerTestRepository{}
	service := services.NewReportManager(repository)

	service.AddMultipageReporter(
		func(c *models.Crawl) *models.MultipageIssueReporter {
			stream := make(chan int64)

			go func() {
				stream <- pageReportId
				close(stream)
			}()

			return &models.MultipageIssueReporter{
				Pstream:   stream,
				ErrorType: reporterErrorType,
			}
		},
	)

	crawl := &models.Crawl{Id: reporterCrawlId}

	service.CreateMultipageIssues(crawl)

	// The repository should contain exactly one issue.
	if len(repository.Issues) != 1 {
		t.Errorf("CreatePageIsssues: %d != 1", len(repository.Issues))
	}

	issue := repository.Issues[0]

	// Make sure the created issue contains the right ids.
	if issue.PageReportId != pageReportId {
		t.Errorf("CreatePageIsssues: PageReportId %d != %d", issue.PageReportId, pageReportId)
	}

	if issue.CrawlId != reporterCrawlId {
		t.Errorf("CreatePageIsssues: reporterCrawlId %d != %d", issue.CrawlId, reporterCrawlId)
	}

	if issue.ErrorType != reporterErrorType {
		t.Errorf("CreatePageIsssues: reporterCrawlId %d != %d", issue.ErrorType, reporterErrorType)
	}
}
