package http

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/stjudewashere/seonaut/internal/helper"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"gopkg.in/yaml.v3"
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

// HTTPServerConfig stores the configuration for the HTTP server.
// It is loaded from the config package.
type HTTPServerConfig struct {
	Server string `mapstructure:"host"`
	Port   int    `mapstructure:"port"`
}

type App struct {
	config         *HTTPServerConfig
	cookie         *sessions.CookieStore
	renderer       *helper.Renderer
	userService    UserService
	projectService ProjectService
	crawlerService CrawlerService
	issueService   IssueService
	reportService  ReportService
	reportManager  ReportManager
}

func NewApp(c *HTTPServerConfig, s *Services) *App {
	translation, err := ioutil.ReadFile("translations/translation.en.yaml")
	if err != nil {
		log.Fatal(err)
	}

	m := make(map[string]interface{})
	err = yaml.Unmarshal(translation, &m)
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
		config:         c,
		cookie:         cookie,
		renderer:       helper.NewRenderer(m),
		userService:    s.UserService,
		projectService: s.ProjectService,
		crawlerService: s.CrawlerService,
		issueService:   s.IssueService,
		reportService:  s.ReportService,
		reportManager:  s.ReportManager,
	}
}

func (app *App) Run() {
	// Static
	fileServer := http.FileServer(http.Dir("./web/static"))
	http.Handle("/resources/", http.StripPrefix("/resources", fileServer))
	http.Handle("/robots.txt", fileServer)
	http.Handle("/favicon.ico", fileServer)

	// App
	http.HandleFunc("/", app.requireAuth(app.serveHome))
	http.HandleFunc("/new-project", app.requireAuth(app.serveProjectAdd))
	http.HandleFunc("/crawl", app.requireAuth(app.serveCrawl))
	http.HandleFunc("/issues", app.requireAuth(app.serveIssues))
	http.HandleFunc("/issues/view", app.requireAuth(app.serveIssuesView))
	http.HandleFunc("/download", app.requireAuth(app.serveDownloadCSV))
	http.HandleFunc("/sitemap", app.requireAuth(app.serveSitemap))
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
