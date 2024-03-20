package container

import (
	"database/sql"
	"log"

	"github.com/stjudewashere/seonaut/internal/cache"
	"github.com/stjudewashere/seonaut/internal/cache_manager"
	"github.com/stjudewashere/seonaut/internal/config"
	"github.com/stjudewashere/seonaut/internal/cookie_session"
	"github.com/stjudewashere/seonaut/internal/crawler_service"
	"github.com/stjudewashere/seonaut/internal/datastore"
	"github.com/stjudewashere/seonaut/internal/export"
	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/project"
	"github.com/stjudewashere/seonaut/internal/projectview"
	"github.com/stjudewashere/seonaut/internal/pubsub"
	"github.com/stjudewashere/seonaut/internal/renderer"
	"github.com/stjudewashere/seonaut/internal/report"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporters"
	"github.com/stjudewashere/seonaut/internal/report_manager/sql_reporters"
	"github.com/stjudewashere/seonaut/internal/user"
)

type Container struct {
	Config             *config.Config
	Datastore          *datastore.Datastore
	PubSubBroker       *pubsub.Broker
	CacheManager       *cache_manager.CacheManager
	IssueService       *issue.Service
	ReportService      *report.Service
	ReportManager      *report_manager.ReportManager
	UserService        *user.Service
	ProjectService     *project.Service
	ProjectViewService *projectview.Service
	ExportService      *export.Exporter
	CrawlerService     *crawler_service.Service
	Renderer           *renderer.Renderer
	CookieSession      *cookie_session.CookieSession

	db    *sql.DB
	cache *cache.MemCache
}

func NewContainer(configFile string) *Container {
	c := &Container{}
	c.InitConfig(configFile)
	c.InitDB()
	c.InitDatastore()
	c.InitPubSubBroker()
	c.InitCache()
	c.InitCacheManager()
	c.InitIssueService()
	c.InitReportService()
	c.InitReportManager()
	c.InitUserService()
	c.InitProjectService()
	c.InitProjectViewService()
	c.InitExportService()
	c.InitCrawlerService()
	c.InitRenderer()
	c.InitCookieSession()

	return c
}

// Load config file using the parameters in configFile.
func (c *Container) InitConfig(configFile string) {
	config, err := config.NewConfig(configFile)
	if err != nil {
		log.Fatalf("Error loading config: %v\n", err)
	}

	c.Config = config
}

// Create the sql database connection.
func (c *Container) InitDB() {
	db, err := datastore.SqlConnect(c.Config.DB)
	if err != nil {
		log.Fatalf("Error creating new database connection: %v\n", err)
	}

	c.db = db
}

// Create database data store.
func (c *Container) InitDatastore() {
	ds, err := datastore.NewDataStore(c.db)
	if err != nil {
		log.Fatalf("Error creating new datastore: %v\n", err)
	}

	c.Datastore = ds
}

// Create the PubSub broker.
func (c *Container) InitPubSubBroker() {
	c.PubSubBroker = pubsub.New()
}

// Create the cache system.
func (c *Container) InitCache() {
	c.cache = cache.NewMemCache()
}

// Create the cache manager.
func (c *Container) InitCacheManager() {
	c.CacheManager = cache_manager.New()
}

// Create the issue service and add it to the cache manager.
func (c *Container) InitIssueService() {
	c.IssueService = issue.NewService(c.Datastore, c.cache)
	c.CacheManager.AddCrawlCacheHandler(c.IssueService)
}

// Create the report service and add it to the cache manager.
func (c *Container) InitReportService() {
	c.ReportService = report.NewService(c.Datastore, c.cache)
	c.CacheManager.AddCrawlCacheHandler(c.ReportService)
}

// Create the report manager and add all the available reporters.
func (c *Container) InitReportManager() {
	c.ReportManager = report_manager.NewReportManager(c.Datastore)
	for _, r := range reporters.GetAllReporters() {
		c.ReportManager.AddPageReporter(r)
	}

	// Create the sql multipage reporters and add them all to the reporterManager.
	sqlReporters := sql_reporters.NewSqlReporter(c.db)
	for _, r := range sqlReporters.GetAllReporters() {
		c.ReportManager.AddMultipageReporter(r)
	}
}

// Create the user service.
func (c *Container) InitUserService() {
	c.UserService = user.NewService(c.Datastore)
}

// Create the Project service.
func (c *Container) InitProjectService() {
	c.ProjectService = project.NewService(c.Datastore, c.CacheManager)
}

// Create the ProjectView service.
func (c *Container) InitProjectViewService() {
	c.ProjectViewService = projectview.NewService(c.Datastore)
}

// Create the Export service.
func (c *Container) InitExportService() {
	c.ExportService = export.NewExporter(c.Datastore)
}

// Create Crawler service.
func (c *Container) InitCrawlerService() {
	crawlerServices := crawler_service.Services{
		Broker:        c.PubSubBroker,
		CacheManager:  c.CacheManager,
		ReportManager: c.ReportManager,
		IssueService:  c.IssueService,
	}

	c.CrawlerService = crawler_service.NewService(c.Datastore, c.Config.Crawler, crawlerServices)
}

// Create html renderer.
func (c *Container) InitRenderer() {
	renderer, err := renderer.NewRenderer(&renderer.RendererConfig{
		TemplatesFolder:  "web/templates",
		TranslationsFile: "translations/translation.en.yaml",
	})
	if err != nil {
		log.Fatal(err)
	}

	c.Renderer = renderer
}

// Create cookie session handler
func (c *Container) InitCookieSession() {
	c.CookieSession = cookie_session.NewCookieSession(c.UserService)
}
