package services

import (
	"log"
	"time"

	"github.com/stjudewashere/seonaut/internal/models"
)

// Available crawl schedule intervals.
const (
	ScheduleHourly = "hourly"
	ScheduleDaily  = "daily"
	ScheduleWeekly = "weekly"
)

// SchedulerTick is how often the scheduler checks for due projects.
const SchedulerTick = time.Minute

// ScheduleIntervals maps the schedule interval names to their duration.
var ScheduleIntervals = map[string]time.Duration{
	ScheduleHourly: time.Hour,
	ScheduleDaily:  24 * time.Hour,
	ScheduleWeekly: 7 * 24 * time.Hour,
}

type SchedulerRepository interface {
	FindScheduledProjects() []models.Project
	UpdateProjectNextRun(p *models.Project, nextRun time.Time) error
}

type CrawlStarter interface {
	StartCrawler(p models.Project, b models.BasicAuth) error
}

// SchedulerService starts crawls for projects with a schedule.
type SchedulerService struct {
	repository   SchedulerRepository
	crawlStarter CrawlStarter
	now          func() time.Time
}

func NewSchedulerService(r SchedulerRepository, c CrawlStarter) *SchedulerService {
	return &SchedulerService{
		repository:   r,
		crawlStarter: c,
		now:          time.Now,
	}
}

// Start runs the scheduler loop until the stop channel is closed.
// On every tick it checks all the scheduled projects and starts a new
// crawl for the due ones.
func (s *SchedulerService) Start(stop <-chan struct{}) {
	ticker := time.NewTicker(SchedulerTick)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			s.RunOnce()
		}
	}
}

// RunOnce starts a crawl for every scheduled project that is due.
// The next run time is always advanced, even when the crawl cannot be
// started, so a failing project does not retry on every tick.
func (s *SchedulerService) RunOnce() {
	for _, p := range s.repository.FindScheduledProjects() {
		interval, ok := ScheduleIntervals[p.ScheduleInterval]
		if !ok {
			continue
		}

		if p.NextRun != nil && p.NextRun.After(s.now()) {
			continue
		}

		if p.BasicAuth {
			// BasicAuth credentials are only provided interactively, they
			// are not stored, so projects using BasicAuth cannot be scheduled.
			log.Printf("Scheduler: skipping project %d (%s), it uses BasicAuth", p.Id, p.URL)
		} else {
			p.Scheduled = true
			if err := s.crawlStarter.StartCrawler(p, models.BasicAuth{}); err != nil {
				log.Printf("Scheduler: error starting crawl for project %d (%s): %v", p.Id, p.URL, err)
			}
		}

		s.repository.UpdateProjectNextRun(&p, s.now().Add(interval))
	}
}
