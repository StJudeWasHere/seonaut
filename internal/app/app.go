package app

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/mnlg/lenkrr/internal/config"
	"github.com/mnlg/lenkrr/internal/crawler"
	"github.com/mnlg/lenkrr/internal/issue"
	"github.com/mnlg/lenkrr/internal/project"
	stripeService "github.com/mnlg/lenkrr/internal/stripe"
	"github.com/mnlg/lenkrr/internal/user"

	"github.com/gorilla/sessions"
	"github.com/microcosm-cc/bluemonday"
	"github.com/stripe/stripe-go/v72"
	"gopkg.in/yaml.v3"
)

type UserService interface {
	Exists(email string) bool
	FindById(id int) *user.User
	SignUp(email, password string) error
	SignIn(email, password string) (*user.User, error)
}

type StripeService interface {
	SetSession(userID int, sessionID string)
	HandleEvent(string, map[string]interface{})
}

type ProjectService interface {
	GetProjects(int) []project.Project
	SaveProject(string, bool, bool, int)
	FindProject(id, uid int) (project.Project, error)
}

type CrawlerService interface {
	StartCrawler(project.Project, string, bool, *bluemonday.Policy) int
}

type IssueService interface {
	GetIssuesCount(int) *issue.IssueCount
}

type App struct {
	config         *config.Config
	datastore      *datastore
	cookie         *sessions.CookieStore
	sanitizer      *bluemonday.Policy
	renderer       *Renderer
	userService    UserService
	stripeService  StripeService
	projectService ProjectService
	crawlerService CrawlerService
	issueService   IssueService
}

func NewApp(c *config.Config, ds *datastore) *App {
	translation, err := ioutil.ReadFile("translation.en.yaml")
	if err != nil {
		log.Fatal(err)
	}

	m := make(map[string]interface{})
	err = yaml.Unmarshal(translation, &m)
	if err != nil {
		log.Fatal(err)
	}

	return &App{
		config:         c,
		datastore:      ds,
		cookie:         sessions.NewCookieStore([]byte("SESSION_ID")),
		sanitizer:      bluemonday.StrictPolicy(),
		renderer:       NewRenderer(m),
		userService:    user.NewService(ds),
		stripeService:  stripeService.NewService(ds),
		projectService: project.NewService(ds),
		crawlerService: crawler.NewService(ds),
		issueService:   issue.NewService(ds),
	}
}

func (app *App) Run() {
	stripe.Key = app.config.Stripe.Secret

	stripe.SetAppInfo(&stripe.AppInfo{
		Name:    "stripe-samples/checkout-single-subscription",
		Version: "0.0.1",
		URL:     "https://github.com/stripe-samples/checkout-single-subscription",
	})

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

	// Stripe
	http.HandleFunc("/upgrade", app.requireAuth(app.upgrade))
	http.HandleFunc("/create-checkout-session", app.requireAuth(app.handleCreateCheckoutSession))
	http.HandleFunc("/checkout-session", app.requireAuth(app.handleCheckoutSession))
	http.HandleFunc("/config", app.requireAuth(app.handleConfig))
	http.HandleFunc("/manage", app.requireAuth(app.handleManageAccount))
	http.HandleFunc("/canceled", app.requireAuth(app.handleCanceled))
	http.HandleFunc("/customer-portal", app.requireAuth(app.handleCustomerPortal))
	http.HandleFunc("/webhook", app.handleWebhook)

	fmt.Printf("Starting server at %s on port %d...\n", app.config.Server, app.config.ServerPort)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", app.config.Server, app.config.ServerPort), nil)
	if err != nil {
		log.Fatal(err)
	}
}
