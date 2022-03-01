package http

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/mnlg/lenkrr/internal/config"
	"github.com/mnlg/lenkrr/internal/crawler"
	"github.com/mnlg/lenkrr/internal/datastore"
	"github.com/mnlg/lenkrr/internal/issue"
	"github.com/mnlg/lenkrr/internal/project"
	"github.com/mnlg/lenkrr/internal/report"
	stripeService "github.com/mnlg/lenkrr/internal/stripe"
	"github.com/mnlg/lenkrr/internal/user"

	"github.com/gorilla/sessions"
	"github.com/microcosm-cc/bluemonday"
	"github.com/stripe/stripe-go/v72"
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
	SaveProject(string, bool, int)
	FindProject(id, uid int) (project.Project, error)
	GetProjectView(id, uid int) (*project.ProjectView, error)
	GetProjectViews(uid int) []project.ProjectView
}

type CrawlerService interface {
	StartCrawler(project.Project, string, bool, *bluemonday.Policy) int
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

type App struct {
	config         *config.Config
	datastore      *datastore.Datastore
	cookie         *sessions.CookieStore
	sanitizer      *bluemonday.Policy
	renderer       *Renderer
	userService    UserService
	stripeService  StripeService
	projectService ProjectService
	crawlerService CrawlerService
	issueService   IssueService
	reportService  ReportService
	reportManager  ReportManager
}

func NewApp(c *config.Config, ds *datastore.Datastore) *App {
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
		reportService:  report.NewService(ds),
		reportManager:  newReportManager(ds),
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

func newReportManager(ds *datastore.Datastore) *issue.ReportManager {
	rm := issue.NewReportManager(ds)

	rm.AddReporter(ds.Find30xPageReports, Error30x)
	rm.AddReporter(ds.Find40xPageReports, Error40x)
	rm.AddReporter(ds.Find50xPageReports, Error50x)
	rm.AddReporter(ds.FindPageReportsWithDuplicatedTitle, ErrorDuplicatedTitle)
	rm.AddReporter(ds.FindPageReportsWithDuplicatedTitle, ErrorDuplicatedDescription)
	rm.AddReporter(ds.FindPageReportsWithEmptyTitle, ErrorEmptyTitle)
	rm.AddReporter(ds.FindPageReportsWithShortTitle, ErrorShortTitle)
	rm.AddReporter(ds.FindPageReportsWithLongTitle, ErrorLongTitle)
	rm.AddReporter(ds.FindPageReportsWithEmptyDescription, ErrorEmptyDescription)
	rm.AddReporter(ds.FindPageReportsWithShortDescription, ErrorShortDescription)
	rm.AddReporter(ds.FindPageReportsWithLongDescription, ErrorLongDescription)
	rm.AddReporter(ds.FindPageReportsWithLittleContent, ErrorLittleContent)
	rm.AddReporter(ds.FindImagesWithNoAlt, ErrorImagesWithNoAlt)
	rm.AddReporter(ds.FindRedirectChains, ErrorRedirectChain)
	rm.AddReporter(ds.FindPageReportsWithoutH1, ErrorNoH1)
	rm.AddReporter(ds.FindPageReportsWithNoLangAttr, ErrorNoLang)
	rm.AddReporter(ds.FindPageReportsWithHTTPLinks, ErrorHTTPLinks)
	rm.AddReporter(ds.FindMissingHrelangReturnLinks, ErrorHreflangsReturnLink)
	rm.AddReporter(ds.TooManyLinks, ErrorTooManyLinks)
	rm.AddReporter(ds.InternalNoFollowLinks, ErrorInternalNoFollow)
	rm.AddReporter(ds.FindExternalLinkWitoutNoFollow, ErrorExternalWithoutNoFollow)
	rm.AddReporter(ds.FindCanonicalizedToNonCanonical, ErrorCanonicalizedToNonCanonical)
	rm.AddReporter(ds.FindCanonicalizedToNonCanonical, ErrorRedirectLoop)
	rm.AddReporter(ds.FindNotValidHeadingsOrder, ErrorNotValidHeadings)

	return rm
}
