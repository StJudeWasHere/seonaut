package app

import (
	"github.com/mnlg/lenkrr/internal/issue"
	"github.com/mnlg/lenkrr/internal/report"
)

const (
	Error30x = iota + 1
	Error40x
	Error50x
	ErrorDuplicatedTitle
	ErrorDuplicatedDescription
	ErrorEmptyTitle
	ErrorShortTitle
	ErrorLongTitle
	ErrorEmptyDescription
	ErrorShortDescription
	ErrorLongDescription
	ErrorLittleContent
	ErrorImagesWithNoAlt
	ErrorRedirectChain
	ErrorNoH1
	ErrorNoLang
	ErrorHTTPLinks
	ErrorHreflangsReturnLink
	ErrorTooManyLinks
	ErrorInternalNoFollow
	ErrorExternalWithoutNoFollow
	ErrorCanonicalizedToNonCanonical
	ErrorRedirectLoop
	ErrorNotValidHeadings
)

const (
	Critical = iota + 1
	Alert
	Warning
)

type IssueCallback struct {
	Callback  func(int) []report.PageReport
	ErrorType int
}

type ReportManager struct {
	callbacks []IssueCallback
}

func (r *ReportManager) addReporter(c func(int) []report.PageReport, t int) {
	r.callbacks = append(r.callbacks, IssueCallback{Callback: c, ErrorType: t})
}

func (r *ReportManager) createIssues(cid int) []issue.Issue {
	var issues []issue.Issue

	for _, c := range r.callbacks {
		for _, p := range c.Callback(cid) {
			i := issue.Issue{
				PageReportId: p.Id,
				ErrorType:    c.ErrorType,
			}

			issues = append(issues, i)
		}
	}

	return issues
}
