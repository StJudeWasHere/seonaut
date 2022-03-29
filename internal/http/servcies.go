package http

import (
	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/project"
	"github.com/stjudewashere/seonaut/internal/projectview"
	"github.com/stjudewashere/seonaut/internal/report"
	"github.com/stjudewashere/seonaut/internal/user"
)

type UserService interface {
	FindById(id int) *user.User
	SignUp(email, password string) error
	SignIn(email, password string) (*user.User, error)
}

type ProjectService interface {
	SaveProject(string, bool, bool, int) error
	FindProject(id, uid int) (project.Project, error)
}

type ProjectViewService interface {
	GetProjectView(id, uid int) (*projectview.ProjectView, error)
	GetProjectViews(uid int) []projectview.ProjectView
}

type CrawlerService interface {
	StartCrawler(project.Project) (int64, error)
}

type IssueService interface {
	GetIssuesCount(int) *issue.IssueCount
	GetPaginatedReportsByIssue(int, int, string) (issue.PaginatorView, error)
}

type ReportService interface {
	GetPageReport(int, int, string) *report.PageReportView
	GetPageReporsByIssueType(int, string) []crawler.PageReport
	GetSitemapPageReports(int) []crawler.PageReport
}

type ReportManager interface {
	CreateIssues(int64) []issue.Issue
}

// Services stores all the services needed by the HTTP server.
type Services struct {
	UserService        UserService
	ProjectService     ProjectService
	ProjectViewService ProjectViewService
	CrawlerService     CrawlerService
	IssueService       IssueService
	ReportService      ReportService
	ReportManager      ReportManager
}
