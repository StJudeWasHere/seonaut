package http

import (
	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/project"
	"github.com/stjudewashere/seonaut/internal/report"
	"github.com/stjudewashere/seonaut/internal/user"

	"github.com/microcosm-cc/bluemonday"
)

type UserService interface {
	FindById(id int) *user.User
	SignUp(email, password string) error
	SignIn(email, password string) (*user.User, error)
}

type ProjectService interface {
	GetProjects(int) []project.Project
	SaveProject(string, bool, bool, int) error
	FindProject(id, uid int) (project.Project, error)
	GetProjectView(id, uid int) (*project.ProjectView, error)
	GetProjectViews(uid int) []project.ProjectView
}

type CrawlerService interface {
	StartCrawler(project.Project, *bluemonday.Policy) int
}

type IssueService interface {
	GetIssuesCount(int) *issue.IssueCount
	GetPaginatedReportsByIssue(int, int, string) (issue.PaginatorView, error)
}

type ReportService interface {
	GetPageReport(int, int, string) *report.PageReportView
	GetPageReporsByIssueType(int, string) []report.PageReport
	GetSitemapPageReports(int) []report.PageReport
}

type ReportManager interface {
	CreateIssues(int) []issue.Issue
}

type Services struct {
	UserService    UserService
	ProjectService ProjectService
	CrawlerService CrawlerService
	IssueService   IssueService
	ReportService  ReportService
	ReportManager  ReportManager
}
