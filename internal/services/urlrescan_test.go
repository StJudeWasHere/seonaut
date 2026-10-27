package services_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stjudewashere/seonaut/internal/config"
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/services"
)

// urlScanTestRepository tracks repo calls for ScanURL verification.
type urlScanTestRepository struct {
	mu            sync.Mutex
	saveCrawlCount int
	updateCount   int
}

func (r *urlScanTestRepository) SaveCrawl(p models.Project) (*models.Crawl, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.saveCrawlCount++
	return &models.Crawl{Id: 7, ProjectId: p.Id, URL: p.URL}, nil
}

func (r *urlScanTestRepository) GetLastCrawl(p *models.Project) models.Crawl {
	return models.Crawl{Id: 7, ProjectId: p.Id, URL: p.URL}
}

func (r *urlScanTestRepository) GetLastCrawls(p models.Project, limit int) []models.Crawl {
	return []models.Crawl{}
}

func (r *urlScanTestRepository) DeleteCrawlData(c *models.Crawl) {}

func (r *urlScanTestRepository) CountIssuesByPriority(crawlId int64, priority int) int { return 0 }

func (r *urlScanTestRepository) UpdateCrawl(c *models.Crawl) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.updateCount++
}

func (r *urlScanTestRepository) snapshot() (int, int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.saveCrawlCount, r.updateCount
}

type urlScanHandlerRepository struct{}

func (r *urlScanHandlerRepository) SavePageReport(pr *models.PageReport, crawlId int64) (*models.PageReport, error) {
	return pr, nil
}

type urlScanReportManagerRepository struct{}

func (r *urlScanReportManagerRepository) SaveIssues(issues <-chan *models.Issue) {
	for range issues {
	}
}

func newURLScanService(repo *urlScanTestRepository) *services.CrawlerService {
	broker := services.NewPubSubBroker()
	reportManager := services.NewReportManager(&urlScanReportManagerRepository{})
	handler := services.NewCrawlerHandler(&urlScanHandlerRepository{}, broker, reportManager)

	return services.NewCrawlerService(repo, services.CrawlerServicesContainer{
		Broker:         broker,
		ReportManager:  reportManager,
		CrawlerHandler: handler,
		ArchiveService: services.NewArchiveService(""),
		Config:         &config.CrawlerConfig{Agent: "testbot"},
	})
}

// TestScanURLRejectsForeignHost verifies ScanURL refuses a URL whose host
// does not match the project's host — no repository writes may happen.
func TestScanURLRejectsForeignHost(t *testing.T) {
	repo := &urlScanTestRepository{}
	svc := newURLScanService(repo)

	p := models.Project{Id: 1, URL: "http://localhost:1/"}

	err := svc.ScanURL(p, "https://evil.example.com/page")
	if err == nil {
		t.Fatal("ScanURL with foreign host: expected error, got nil")
	}

	sc, uc := repo.snapshot()
	if sc != 0 || uc != 0 {
		t.Errorf("foreign host scan must not touch the DB: SaveCrawl=%d UpdateCrawl=%d", sc, uc)
	}
}

// TestScanURLStartsCrawl verifies ScanURL creates a crawl record and starts
// the crawl on the target URL.
func TestScanURLStartsCrawl(t *testing.T) {
	repo := &urlScanTestRepository{}
	svc := newURLScanService(repo)

	p := models.Project{Id: 1, URL: "http://localhost:1/"}

	err := svc.ScanURL(p, "http://localhost:1/some-page")
	if err != nil {
		t.Fatalf("ScanURL: unexpected error: %v", err)
	}

	// localhost:1 refuses connections, so the background crawl goroutine
	// finishes quickly like TestStartCrawlerNoDuplicateDBRecord. Wait for
	// UpdateCrawl to be called.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		_, uc := repo.snapshot()
		if uc >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	sc, uc := repo.snapshot()
	if sc != 1 {
		t.Errorf("SaveCrawl called %d time(s), want 1", sc)
	}
	if uc != 1 {
		t.Errorf("UpdateCrawl called %d time(s), want 1", uc)
	}
}

// TestScanURLLockedProject verifies ScanURL respects the in-memory crawl lock
// and returns an error while a crawl is already running.
func TestScanURLLockedProject(t *testing.T) {
	repo := &urlScanTestRepository{}
	svc := newURLScanService(repo)

	p := models.Project{Id: 1, URL: "http://localhost:1/"}

	if err := svc.ScanURL(p, "http://localhost:1/a"); err != nil {
		t.Fatalf("first ScanURL: unexpected error: %v", err)
	}

	err := svc.ScanURL(p, "http://localhost:1/b")
	if err == nil {
		t.Fatal("second ScanURL: expected error while crawl is in progress, got nil")
	}
}
