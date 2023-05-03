package report_manager_test

import (
	"testing"
	"time"

	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporters"
)

const (
	crawlId      = 1
	pageReportId = 1
	errorId      = 1
)

type storage struct{}

func (s *storage) SaveIssues(c <-chan *issue.Issue) {
	<-c
}
func (s *storage) SaveEndIssues(crawlId int64, t time.Time) {}
func (s *storage) PageReportsQuery(query string, args ...interface{}) <-chan *models.PageReport {
	prStream := make(chan *models.PageReport)
	return prStream
}

var service = report_manager.NewReportManager(&storage{})

func TestCreateIssues(t *testing.T) {
	pageReport := &models.PageReport{Id: pageReportId}
	total := 0

	service.AddPageReporter(
		&reporters.PageIssueReporter{
			ErrorType: 1,
			Callback: func(pageReport *models.PageReport) bool {
				total = 1
				return true
			},
		})

	service.CreatePageIssues(pageReport, &models.Crawl{})
	if total != 1 {
		t.Errorf("CreateIsssues: %d != 1", total)
	}
}
