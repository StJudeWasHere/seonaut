package issue

import (
	"time"

	"github.com/stjudewashere/seonaut/internal/crawler"
)

type IssueCallback struct {
	Callback  func(int) []crawler.PageReport
	ErrorType int
}

type ReportManager struct {
	store     ReportManagerStore
	callbacks []IssueCallback
}

type ReportManagerStore interface {
	SaveIssues([]Issue, int)
	SaveEndIssues(int, time.Time, int)
}

func NewReportManager(s ReportManagerStore) *ReportManager {
	return &ReportManager{
		store: s,
	}
}

func (r *ReportManager) AddReporter(c func(int) []crawler.PageReport, t int) {
	r.callbacks = append(r.callbacks, IssueCallback{Callback: c, ErrorType: t})
}

func (r *ReportManager) CreateIssues(cid int) []Issue {
	var issues []Issue

	for _, c := range r.callbacks {
		for _, p := range c.Callback(cid) {
			i := Issue{
				PageReportId: p.Id,
				ErrorType:    c.ErrorType,
			}

			issues = append(issues, i)
		}
	}

	r.store.SaveIssues(issues, cid)
	r.store.SaveEndIssues(cid, time.Now(), len(issues))

	return issues
}
