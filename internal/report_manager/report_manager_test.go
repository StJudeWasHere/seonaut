package report_manager_test

import (
	"testing"
	"time"

	"github.com/stjudewashere/seonaut/internal/cache_manager"
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

var service = report_manager.NewReportManager(&storage{}, cache_manager.New())

func TestCreateIssues(t *testing.T) {
	pageReports := []models.PageReport{
		{Id: pageReportId},
	}

	total := 0

	service.AddReporter(func(reporters.DatabaseReporter, *models.Crawl) <-chan *models.PageReport {
		prStream := make(chan *models.PageReport)
		go func() {
			defer close(prStream)
			for _, v := range pageReports {
				prStream <- &v
				total++
			}
		}()
		return prStream
	}, errorId)

	service.CreateIssues(&models.Crawl{Id: crawlId})
	if total != 1 {
		t.Errorf("CreateIsssues: %d != 1", total)
	}
}
