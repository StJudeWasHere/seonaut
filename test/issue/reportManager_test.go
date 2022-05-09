package issue

import (
	"testing"
	"time"

	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/issue"
)

const (
	crawlId      = 1
	pageReportId = 1
	errorId      = 1
)

type storage struct{}

func (s *storage) SaveIssues(issues []issue.Issue, crawlId int64)      {}
func (s *storage) SaveEndIssues(crawlId int64, t time.Time, total int) {}

var service = issue.NewReportManager(&storage{})

func TestCreateIssues(t *testing.T) {
	pageReports := []crawler.PageReport{
		crawler.PageReport{Id: pageReportId},
	}

	service.AddReporter(func(crawlId int64) []crawler.PageReport {
		return pageReports
	}, errorId)

	issues := service.CreateIssues(crawlId)
	if len(issues) != 1 {
		t.Errorf("CreateIsssues: %d != 1", len(issues))
	}

	if issues[0].PageReportId != pageReportId {
		t.Errorf("pageReport id: %d != %d", issues[0].PageReportId, pageReportId)
	}

	if issues[0].ErrorType != errorId {
		t.Errorf("errorType id: %d != %d", issues[0].ErrorType, errorId)
	}
}
