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
	SaveProject(string, bool, bool, bool, int) error
	FindProject(id, uid int) (project.Project, error)
}

type ProjectViewService interface {
	GetProjectView(id, uid int) (*projectview.ProjectView, error)
	GetProjectViews(uid int) []projectview.ProjectView
}

type CrawlerService interface {
	StartCrawler(project.Project) (*crawler.Crawl, error)
}

type IssueService interface {
	GetIssuesCount(int64) *issue.IssueCount
	GetPaginatedReportsByIssue(int64, int, string) (issue.PaginatorView, error)
}

type ReportService interface {
	GetPageReport(int, int64, string) *report.PageReportView
	GetPageReporsByIssueType(int64, string) []crawler.PageReport
	GetSitemapPageReports(int64) []crawler.PageReport
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
