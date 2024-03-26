package container

import (
	"database/sql"
	"log"

	"github.com/stjudewashere/seonaut/internal/config"
	"github.com/stjudewashere/seonaut/internal/datastore"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporters"
	"github.com/stjudewashere/seonaut/internal/report_manager/sql_reporters"
)

type Container struct {
	Config             *config.Config
	Datastore          *datastore.Datastore
	PubSubBroker       *Broker
	CacheManager       *CacheManager
	IssueService       *IssueService
	ReportService      *ReportService
	ReportManager      *report_manager.ReportManager
	UserService        *UserService
	ProjectService     *ProjectService
	ProjectViewService *ProjectViewService
	ExportService      *Exporter
	CrawlerService     *CrawlerService
	Renderer           *Renderer
	CookieSession      *CookieSession

	db    *sql.DB
	cache *MemCache
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
	c.PubSubBroker = NewPubSubBroker()
}

// Create the cache system.
func (c *Container) InitCache() {
	c.cache = NewMemCache()
}

// Create the cache manager.
func (c *Container) InitCacheManager() {
	c.CacheManager = NewCacheManager()
}

// Create the issue service and add it to the cache manager.
func (c *Container) InitIssueService() {
	c.IssueService = NewIssueService(c.Datastore, c.cache)
	c.CacheManager.AddCrawlCacheHandler(c.IssueService)
}

// Create the report service and add it to the cache manager.
func (c *Container) InitReportService() {
	c.ReportService = NewReportService(c.Datastore, c.cache)
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
	c.UserService = NewUserService(c.Datastore)
}

// Create the Project service.
func (c *Container) InitProjectService() {
	c.ProjectService = NewProjectService(c.Datastore, c.CacheManager)
}

// Create the ProjectView service.
func (c *Container) InitProjectViewService() {
	c.ProjectViewService = NewProjectViewService(c.Datastore)
}

// Create the Export service.
func (c *Container) InitExportService() {
	c.ExportService = NewExporter(c.Datastore)
}

// Create Crawler service.
func (c *Container) InitCrawlerService() {
	crawlerServices := Services{
		Broker:        c.PubSubBroker,
		CacheManager:  c.CacheManager,
		ReportManager: c.ReportManager,
		IssueService:  c.IssueService,
	}

	c.CrawlerService = NewCrawlerService(c.Datastore, c.Config.Crawler, crawlerServices)
}

// Create html renderer.
func (c *Container) InitRenderer() {
	renderer, err := NewRenderer(&RendererConfig{
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
	c.CookieSession = NewCookieSession(c.UserService)
}
