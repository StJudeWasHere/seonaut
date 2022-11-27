package issue_test

import (
	"testing"
	"time"

	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/pagereport"
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
func (s *storage) SaveEndIssues(crawlId int64, t time.Time, total int) {}

var service = issue.NewReportManager(&storage{})

func TestCreateIssues(t *testing.T) {
	pageReports := []pagereport.PageReport{
		{Id: pageReportId},
	}

	total := 0

	service.AddReporter(func(crawlId int64) <-chan *pagereport.PageReport {
		prStream := make(chan *pagereport.PageReport)
		go func() {
			defer close(prStream)
			for _, v := range pageReports {
				prStream <- &v
				total++
			}
		}()
		return prStream
	}, errorId)

	service.CreateIssues(crawlId)
	if total != 1 {
		t.Errorf("CreateIsssues: %d != 1", total)
	}
}
