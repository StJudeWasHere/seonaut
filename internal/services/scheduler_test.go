package services

import (
	"errors"
	"testing"
	"time"

	"github.com/stjudewashere/seonaut/internal/models"
)

type schedulerTestRepository struct {
	projects []models.Project
	updated  map[int64]time.Time
}

func (r *schedulerTestRepository) FindScheduledProjects() []models.Project {
	return r.projects
}

func (r *schedulerTestRepository) UpdateProjectNextRun(p *models.Project, nextRun time.Time) error {
	r.updated[p.Id] = nextRun
	return nil
}

type schedulerTestCrawlStarter struct {
	started []int64
	errOn   int64
}

func (c *schedulerTestCrawlStarter) StartCrawler(p models.Project, b models.BasicAuth) error {
	c.started = append(c.started, p.Id)
	if p.Id == c.errOn {
		return errors.New("already crawling")
	}

	return nil
}

func newSchedulerTest(t *testing.T, projects []models.Project) (*SchedulerService, *schedulerTestRepository, *schedulerTestCrawlStarter) {
	t.Helper()

	repo := &schedulerTestRepository{projects: projects, updated: map[int64]time.Time{}}
	starter := &schedulerTestCrawlStarter{}
	s := NewSchedulerService(repo, starter)
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }

	return s, repo, starter
}

func TestRunOnceStartsDueProjects(t *testing.T) {
	past := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)
	future := time.Date(2026, 10, 9, 14, 0, 0, 0, time.UTC)
	s, repo, starter := newSchedulerTest(t, []models.Project{
		{Id: 1, URL: "https://due-nil.example", ScheduleInterval: ScheduleHourly},
		{Id: 2, URL: "https://due-past.example", ScheduleInterval: ScheduleDaily, NextRun: &past},
		{Id: 3, URL: "https://not-due.example", ScheduleInterval: ScheduleDaily, NextRun: &future},
		{Id: 4, URL: "https://bogus.example", ScheduleInterval: "fortnightly", NextRun: &past},
	})

	s.RunOnce()

	if len(starter.started) != 2 || starter.started[0] != 1 || starter.started[1] != 2 {
		t.Fatalf("expected crawls started for projects 1 and 2, got %v", starter.started)
	}

	if got := repo.updated[1].Sub(s.now()); got != time.Hour {
		t.Fatalf("project 1 next_run = now+%v, want +1h", got)
	}

	if got := repo.updated[2].Sub(s.now()); got != 24*time.Hour {
		t.Fatalf("project 2 next_run = now+%v, want +24h", got)
	}

	if _, ok := repo.updated[3]; ok {
		t.Fatal("not-due project must not advance next_run")
	}

	if _, ok := repo.updated[4]; ok {
		t.Fatal("bogus interval must be ignored")
	}
}

func TestRunOnceAdvancesNextRunOnError(t *testing.T) {
	past := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)
	s, repo, _ := newSchedulerTest(t, []models.Project{
		{Id: 7, URL: "https://busy.example", ScheduleInterval: ScheduleWeekly, NextRun: &past},
	})

	s.crawlStarter.(*schedulerTestCrawlStarter).errOn = 7
	s.RunOnce()

	if _, ok := repo.updated[7]; !ok {
		t.Fatal("failing project must still advance next_run to avoid retry storm")
	}
}

func TestRunOnceSkipsBasicAuthProjects(t *testing.T) {
	s, repo, starter := newSchedulerTest(t, []models.Project{
		{Id: 9, URL: "https://auth.example", ScheduleInterval: ScheduleHourly, BasicAuth: true},
	})

	s.RunOnce()

	if len(starter.started) != 0 {
		t.Fatalf("BasicAuth project must not start, got %v", starter.started)
	}

	if _, ok := repo.updated[9]; !ok {
		t.Fatal("BasicAuth project must still advance next_run to avoid tick spam")
	}
}
