package http

import (
	"context"
	"io"

	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/pagereport"
	"github.com/stjudewashere/seonaut/internal/project"
	"github.com/stjudewashere/seonaut/internal/projectview"
	"github.com/stjudewashere/seonaut/internal/pubsub"
	"github.com/stjudewashere/seonaut/internal/report"
	"github.com/stjudewashere/seonaut/internal/user"
)

type UserService interface {
	FindById(id int) *user.User
	SignUp(email, password string) error
	SignIn(email, password string) (*user.User, error)
	UpdatePassword(email, password string) error
	GetUserFromContext(c context.Context) (*user.User, bool)
	SetUserToContext(user *user.User, c context.Context) context.Context
}

type ProjectService interface {
	SaveProject(*project.Project, int) error
	DeleteProject(p *project.Project)
	FindProject(id, uid int) (project.Project, error)
	UpdateProject(p *project.Project) error
}

type ProjectViewService interface {
	GetProjectView(id, uid int) (*projectview.ProjectView, error)
	GetProjectViews(uid int) []projectview.ProjectView
}

type CrawlerService interface {
	StartCrawler(project.Project) (*crawler.Crawl, error)
	GetLastCrawls(p project.Project) []crawler.Crawl
}

type IssueService interface {
	GetIssuesCount(int64) *issue.IssueCount
	GetPaginatedReportsByIssue(int64, int, string) (issue.PaginatorView, error)
	GetLinksCount(int64) *issue.LinksCount
	SaveCrawlIssuesCount(int64)
	GetCanonicalCount(crawlId int64) *issue.CanonicalCount
	GetImageAltCount(crawlId int64) *issue.AltCount
	GetSchemeCount(crawlId int64) *issue.SchemeCount
}

type ReportService interface {
	GetPageReport(int, int64, string, int) *report.PageReportView
	GetPageReporsByIssueType(int64, string) <-chan *pagereport.PageReport
	GetSitemapPageReports(int64) <-chan *pagereport.PageReport
}

type ReportManager interface {
	CreateIssues(int64)
}

type PubSubBroker interface {
	NewSubscriber(topic string, c func(*pubsub.Message) error) *pubsub.Subscriber
	Publish(topic string, m *pubsub.Message)
	Unsubscribe(s *pubsub.Subscriber)
}

type Exporter interface {
	ExportLinks(f io.Writer, crawl *crawler.Crawl)
	ExportExternalLinks(f io.Writer, crawl *crawler.Crawl)
	ExportImages(f io.Writer, crawl *crawler.Crawl)
	ExportScripts(f io.Writer, crawl *crawler.Crawl)
	ExportStyles(f io.Writer, crawl *crawler.Crawl)
	ExportIframes(f io.Writer, crawl *crawler.Crawl)
	ExportAudios(f io.Writer, crawl *crawler.Crawl)
	ExportVideos(f io.Writer, crawl *crawler.Crawl)
	ExportHreflangs(f io.Writer, crawl *crawler.Crawl)
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
	PubSubBroker       PubSubBroker
	ExportService      Exporter
}
