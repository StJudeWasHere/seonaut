package http

import (
	"fmt"
	"log"
	"net/http"

	"github.com/stjudewashere/seonaut/internal/crawler_service"
	"github.com/stjudewashere/seonaut/internal/export"
	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/project"
	"github.com/stjudewashere/seonaut/internal/projectview"
	"github.com/stjudewashere/seonaut/internal/pubsub"
	"github.com/stjudewashere/seonaut/internal/report"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/user"
)

// HTTPServerConfig stores the configuration for the HTTP server.
// It is loaded from the config package.
type HTTPServerConfig struct {
	Server string `mapstructure:"host"`
	Port   int    `mapstructure:"port"`
	URL    string `mapstructure:"url"`
}

// Services stores all the services needed by the HTTP server.
type Services struct {
	UserService        *user.Service
	ProjectService     *project.Service
	ProjectViewService *projectview.Service
	CrawlerService     *crawler_service.Service
	IssueService       *issue.Service
	ReportService      *report.Service
	ReportManager      *report_manager.ReportManager
	PubSubBroker       *pubsub.Broker
	ExportService      *export.Exporter
}

// App is the server application, and it contains all the needed services to handle requests.
type App struct {
	config             *HTTPServerConfig
	cookieSession      *CookieSession
	renderer           *Renderer
	userService        *user.Service
	projectService     *project.Service
	crawlerService     *crawler_service.Service
	issueService       *issue.Service
	reportService      *report.Service
	projectViewService *projectview.Service
	pubsubBroker       *pubsub.Broker
	exportService      *export.Exporter
}

// PageView is the data structure used to render the html templates.
type PageView struct {
	PageTitle string
	User      models.User
	Data      interface{}
	Refresh   bool
}

// NewApp initializes the template renderer and the session cookie.
// Returns a new HTTP application server.
func NewApp(c *HTTPServerConfig, s *Services) *App {
	renderer, err := NewRenderer(&RendererConfig{
		TemplatesFolder:  "web/templates",
		TranslationsFile: "translations/translation.en.yaml",
	})
	if err != nil {
		log.Fatal(err)
	}

	cookieSession := NewCookieSession(s.UserService)

	return &App{
		config:             c,
		cookieSession:      cookieSession,
		renderer:           renderer,
		userService:        s.UserService,
		projectService:     s.ProjectService,
		crawlerService:     s.CrawlerService,
		issueService:       s.IssueService,
		reportService:      s.ReportService,
		projectViewService: s.ProjectViewService,
		pubsubBroker:       s.PubSubBroker,
		exportService:      s.ExportService,
	}
}

// Start HTTP server at the server and port specified in the config.
func (app *App) Run() {
	// Static
	fileServer := http.FileServer(http.Dir("./web/static"))
	http.Handle("/resources/", http.StripPrefix("/resources", fileServer))
	http.Handle("/robots.txt", fileServer)
	http.Handle("/favicon.ico", fileServer)

	// App
	http.HandleFunc("/", app.cookieSession.Auth(app.handleHome))
	http.HandleFunc("/new-project", app.cookieSession.Auth(app.handleProjectAdd))
	http.HandleFunc("/edit-project", app.cookieSession.Auth(app.handleProjectEdit))
	http.HandleFunc("/delete-project", app.cookieSession.Auth(app.handleDeleteProject))
	http.HandleFunc("/crawl", app.cookieSession.Auth(app.handleCrawl))
	http.HandleFunc("/crawl-stop", app.cookieSession.Auth(app.handleStopCrawl))
	http.HandleFunc("/crawl-live", app.cookieSession.Auth(app.handleCrawlLive))
	http.HandleFunc("/crawl-auth", app.cookieSession.Auth(app.handleCrawlAuth))
	http.HandleFunc("/crawl-ws", app.cookieSession.Auth(app.handleCrawlWs))
	http.HandleFunc("/issues", app.cookieSession.Auth(app.handleIssues))
	http.HandleFunc("/issues/view", app.cookieSession.Auth(app.handleIssuesView))
	http.HandleFunc("/dashboard", app.cookieSession.Auth(app.handleDashboard))
	http.HandleFunc("/download", app.cookieSession.Auth(app.handleDownloadCSV))
	http.HandleFunc("/sitemap", app.cookieSession.Auth(app.handleSitemap))
	http.HandleFunc("/export", app.cookieSession.Auth(app.handleExport))
	http.HandleFunc("/export/download", app.cookieSession.Auth(app.handleExportResources))
	http.HandleFunc("/resources", app.cookieSession.Auth(app.handleResourcesView))
	http.HandleFunc("/signout", app.cookieSession.Auth(app.handleSignout))
	http.HandleFunc("/account", app.cookieSession.Auth(app.handleAccount))
	http.HandleFunc("/delete-account", app.cookieSession.Auth((app.handleDeleteUser)))
	http.HandleFunc("/explorer", app.cookieSession.Auth(app.handleExplorer))
	http.HandleFunc("/signup", app.handleSignup)
	http.HandleFunc("/signin", app.handleSignin)

	fmt.Printf("Starting server at %s on port %d...\n", app.config.Server, app.config.Port)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", app.config.Server, app.config.Port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
