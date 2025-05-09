package services

import (
	"database/sql"
	"log"

	"github.com/stjudewashere/seonaut/internal/config"
	"github.com/stjudewashere/seonaut/internal/issues/multipage"
	"github.com/stjudewashere/seonaut/internal/issues/page"
	"github.com/stjudewashere/seonaut/internal/repository"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Container struct {
	Config             *config.Config
	PubSubBroker       *Broker
	IssueService       *IssueService
	ReportService      *ReportService
	ReportManager      *ReportManager
	UserService        *UserService
	DashboardService   *DashboardService
	ProjectService     *ProjectService
	ProjectViewService *ProjectViewService
	ExportService      *Exporter
	CrawlerService     *CrawlerService
	Translator         *Translator
	Renderer           *Renderer
	CookieSession      *CookieSession
	ArchiveService     *ArchiveService
	ReplayService      *ReplayService

	db                   *sql.DB
	issueRepository      *repository.IssueRepository
	pageReportRepository *repository.PageReportRepository
	userRepository       *repository.UserRepository
	projectRepository    *repository.ProjectRepository
	exportRepository     *repository.ExportRepository
	crawlRepository      *repository.CrawlRepository
	dashboardRepository  *repository.DashboardRepository
}

func NewContainer(configFile string) *Container {
	c := &Container{}
	c.InitConfig(configFile)
	c.InitDB()
	c.InitArchiveService()
	c.InitRepositories()
	c.InitPubSubBroker()
	c.InitIssueService()
	c.InitReportService()
	c.InitReportManager()
	c.InitUserService()
	c.InitDashboardService()
	c.InitProjectService()
	c.InitProjectViewService()
	c.InitTranslator()
	c.InitExportService()
	c.InitCrawlerService()
	c.InitRenderer()
	c.InitCookieSession()
	c.InitReplayService()

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

// Create the sql database connection and run migrations.
func (c *Container) InitDB() {
	db, err := repository.SqlConnect(c.Config.DB)
	if err != nil {
		log.Fatalf("Error creating new database connection: %v", err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatalf("Error creating mysql driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "mysql", driver)
	if err != nil {
		log.Fatalf("Error with mysql migrations: %v", err)
	}

	m.Up()

	c.db = db
}

// Create the data repositories.
func (c *Container) InitRepositories() {
	c.issueRepository = &repository.IssueRepository{DB: c.db}
	c.pageReportRepository = &repository.PageReportRepository{DB: c.db}
	c.userRepository = &repository.UserRepository{DB: c.db}
	c.projectRepository = &repository.ProjectRepository{DB: c.db}
	c.exportRepository = &repository.ExportRepository{DB: c.db}
	c.crawlRepository = &repository.CrawlRepository{DB: c.db}
	c.dashboardRepository = &repository.DashboardRepository{DB: c.db}

	// Clean up unfinished crawls.
	c.crawlRepository.DeleteUnfinishedCrawls()
}

// Create the PubSub broker.
func (c *Container) InitPubSubBroker() {
	c.PubSubBroker = NewPubSubBroker()
}

// Create the issue service.
func (c *Container) InitIssueService() {
	c.IssueService = NewIssueService(c.issueRepository)
}

// Create the report service.
func (c *Container) InitReportService() {
	repository := &struct {
		*repository.PageReportRepository
		*repository.IssueRepository
	}{
		c.pageReportRepository,
		c.issueRepository,
	}

	c.ReportService = NewReportService(repository)
}

// Create the report manager and add all the available reporters.
func (c *Container) InitReportManager() {
	c.ReportManager = NewReportManager(c.issueRepository)
	for _, r := range page.GetAllReporters() {
		c.ReportManager.AddPageReporter(r)
	}

	// Create the sql multipage reporters and add them all to the reporterManager.
	sqlReporters := multipage.NewSqlReporter(c.db)
	for _, r := range sqlReporters.GetAllReporters() {
		c.ReportManager.AddMultipageReporter(r)
	}
}

// Create the user service.
func (c *Container) InitUserService() {
	c.UserService = NewUserService(c.userRepository)
}

// Create the Project service.
func (c *Container) InitProjectService() {
	repository := &struct {
		*repository.ProjectRepository
		*repository.CrawlRepository
	}{
		c.projectRepository,
		c.crawlRepository,
	}

	c.ProjectService = NewProjectService(repository, c.ArchiveService)

	// UserService DeleteHooks are called when a user is deleted.
	// Add a DeleteHook so it deletes all user projects and crawl
	// data when a user is deleted.
	c.UserService.AddDeleteHook(c.ProjectService.DeleteAllUserProjects)
}

// Create the ProjectView service.
func (c *Container) InitProjectViewService() {
	repository := &struct {
		*repository.ProjectRepository
		*repository.CrawlRepository
	}{
		c.projectRepository,
		c.crawlRepository,
	}

	c.ProjectViewService = NewProjectViewService(repository)
}

// Create the Export service.
func (c *Container) InitExportService() {
	c.ExportService = NewExporter(c.exportRepository, c.Translator)
}

// Create Crawler service.
func (c *Container) InitCrawlerService() {
	crawlerServices := CrawlerServicesContainer{
		Broker:         c.PubSubBroker,
		ReportManager:  c.ReportManager,
		CrawlerHandler: NewCrawlerHandler(c.pageReportRepository, c.PubSubBroker, c.ReportManager),
		ArchiveService: c.ArchiveService,
		Config:         c.Config.Crawler,
	}
	repository := &struct {
		*repository.CrawlRepository
		*repository.IssueRepository
	}{
		c.crawlRepository,
		c.issueRepository,
	}

	c.CrawlerService = NewCrawlerService(repository, crawlerServices)
}

// Create the dashboCallbackBuilderard service.
func (c *Container) InitDashboardService() {
	c.DashboardService = NewDashboardService(c.dashboardRepository)
}

// Create The translator.
func (c *Container) InitTranslator() {
	var err error
	c.Translator, err = NewTranslator(&TranslatorConfig{
		TranslationsFile: "translations/translation.en.yaml",
	})
	if err != nil {
		log.Fatal(err)
	}
}

// Create html renderer.
func (c *Container) InitRenderer() {
	renderer, err := NewRenderer(&RendererConfig{
		TemplatesFolder: "web/templates",
	}, c.Translator)
	if err != nil {
		log.Fatal(err)
	}

	c.Renderer = renderer
}

// Create cookie session handler
func (c *Container) InitCookieSession() {
	c.CookieSession = NewCookieSession(c.userRepository)
}

// Init the WACZ archiver service.
func (c *Container) InitArchiveService() {
	c.ArchiveService = NewArchiveService("archive")
}

// Init the WACZ archive replay service.
func (c *Container) InitReplayService() {
	c.ReplayService = NewReplayService()
}
