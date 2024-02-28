package http

import (
	"fmt"
	"log"
	"net/http"

	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/export"
	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/project"
	"github.com/stjudewashere/seonaut/internal/projectview"
	"github.com/stjudewashere/seonaut/internal/pubsub"
	"github.com/stjudewashere/seonaut/internal/renderer"
	"github.com/stjudewashere/seonaut/internal/report"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/user"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
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
	CrawlerService     *crawler.Service
	IssueService       *issue.Service
	ReportService      *report.Service
	ReportManager      *report_manager.ReportManager
	PubSubBroker       *pubsub.Broker
	ExportService      *export.Exporter
}

// App is the server application, and it contains all the needed services to handle requests.
type App struct {
	config             *HTTPServerConfig
	cookie             *sessions.CookieStore
	renderer           *renderer.Renderer
	userService        *user.Service
	projectService     *project.Service
	crawlerService     *crawler.Service
	issueService       *issue.Service
	reportService      *report.Service
	reportManager      *report_manager.ReportManager
	projectViewService *projectview.Service
	pubsubBroker       *pubsub.Broker
	exportService      *export.Exporter
}

// PageView is the data structure used to render the html templates.
type PageView struct {
	PageTitle string
	User      user.User
	Data      interface{}
	Refresh   bool
}

// NewApp initializes the template renderer and the session cookie.
// Returns a new HTTP application server.
func NewApp(c *HTTPServerConfig, s *Services) *App {
	renderer, err := renderer.NewRenderer(&renderer.RendererConfig{
		TemplatesFolder:  "web/templates",
		TranslationsFile: "translations/translation.en.yaml",
	})
	if err != nil {
		log.Fatal(err)
	}

	authKeyOne := securecookie.GenerateRandomKey(64)
	encryptionKeyOne := securecookie.GenerateRandomKey(32)

	cookie := sessions.NewCookieStore(
		authKeyOne,
		encryptionKeyOne,
	)

	cookie.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 15,
		HttpOnly: true,
	}

	return &App{
		config:             c,
		cookie:             cookie,
		renderer:           renderer,
		userService:        s.UserService,
		projectService:     s.ProjectService,
		crawlerService:     s.CrawlerService,
		issueService:       s.IssueService,
		reportService:      s.ReportService,
		reportManager:      s.ReportManager,
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
	http.HandleFunc("/", app.requireAuth(app.handleHome))
	http.HandleFunc("/new-project", app.requireAuth(app.handleProjectAdd))
	http.HandleFunc("/edit-project", app.requireAuth(app.handleProjectEdit))
	http.HandleFunc("/delete-project", app.requireAuth(app.handleDeleteProject))
	http.HandleFunc("/crawl", app.requireAuth(app.handleCrawl))
	http.HandleFunc("/crawl-stop", app.requireAuth(app.handleStopCrawl))
	http.HandleFunc("/crawl-live", app.requireAuth(app.handleCrawlLive))
	http.HandleFunc("/crawl-auth", app.requireAuth(app.handleCrawlAuth))
	http.HandleFunc("/crawl-ws", app.requireAuth(app.handleCrawlWs))
	http.HandleFunc("/issues", app.requireAuth(app.handleIssues))
	http.HandleFunc("/issues/view", app.requireAuth(app.handleIssuesView))
	http.HandleFunc("/dashboard", app.requireAuth(app.handleDashboard))
	http.HandleFunc("/download", app.requireAuth(app.handleDownloadCSV))
	http.HandleFunc("/sitemap", app.requireAuth(app.handleSitemap))
	http.HandleFunc("/export", app.requireAuth(app.handleExport))
	http.HandleFunc("/export/download", app.requireAuth(app.handleExportResources))
	http.HandleFunc("/resources", app.requireAuth(app.handleResourcesView))
	http.HandleFunc("/signout", app.requireAuth(app.handleSignout))
	http.HandleFunc("/account", app.requireAuth(app.handleAccount))
	http.HandleFunc("/delete-account", app.requireAuth((app.handleDeleteUser)))
	http.HandleFunc("/explorer", app.requireAuth(app.handleExplorer))
	http.HandleFunc("/signup", app.handleSignup)
	http.HandleFunc("/signin", app.handleSignin)

	fmt.Printf("Starting server at %s on port %d...\n", app.config.Server, app.config.Port)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", app.config.Server, app.config.Port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
