package http

import (
	"fmt"
	"log"
	"net/http"

	"github.com/stjudewashere/seonaut/internal/helper"

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

type App struct {
	config             *HTTPServerConfig
	cookie             *sessions.CookieStore
	renderer           *helper.Renderer
	userService        UserService
	projectService     ProjectService
	crawlerService     CrawlerService
	issueService       IssueService
	reportService      ReportService
	reportManager      ReportManager
	projectViewService ProjectViewService
	pubsubBroker       PubSubBroker
	exportService      Exporter
}

// NewApp initializes the template renderer and the session cookie.
// Returns a new HTTP application server.
func NewApp(c *HTTPServerConfig, s *Services) *App {
	renderer, err := helper.NewRenderer(&helper.RendererConfig{
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
	http.HandleFunc("/", app.requireAuth(app.serveHome))
	http.HandleFunc("/new-project", app.requireAuth(app.serveProjectAdd))
	http.HandleFunc("/delete-project", app.requireAuth(app.serveDeleteProject))
	http.HandleFunc("/crawl", app.requireAuth(app.serveCrawl))
	http.HandleFunc("/crawl-live", app.requireAuth(app.serveCrawlLive))
	http.HandleFunc("/crawl-ws", app.requireAuth(app.serveCrawlWs))
	http.HandleFunc("/issues", app.requireAuth(app.serveIssues))
	http.HandleFunc("/issues/view", app.requireAuth(app.serveIssuesView))
	http.HandleFunc("/dashboard", app.requireAuth(app.serveDashboard))
	http.HandleFunc("/download", app.requireAuth(app.serveDownloadCSV))
	http.HandleFunc("/sitemap", app.requireAuth(app.serveSitemap))
	http.HandleFunc("/export", app.requireAuth(app.serveExport))
	http.HandleFunc("/export/download", app.requireAuth(app.serveExportResources))
	http.HandleFunc("/resources", app.requireAuth(app.serveResourcesView))
	http.HandleFunc("/signout", app.requireAuth(app.serveSignout))
	http.HandleFunc("/signup", app.serveSignup)
	http.HandleFunc("/signin", app.serveSignin)

	fmt.Printf("Starting server at %s on port %d...\n", app.config.Server, app.config.Port)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", app.config.Server, app.config.Port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
