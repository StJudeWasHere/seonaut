package services

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/stjudewashere/seonaut/internal/models"
)

// WebhookTimeout is the timeout used when posting a crawl notification.
const WebhookTimeout = 10 * time.Second

// CrawlNotification is the JSON payload POSTed to a project's webhook URL
// after a crawl finishes.
type CrawlNotification struct {
	ProjectID      int64     `json:"project_id"`
	URL            string    `json:"url"`
	CrawlID        int64     `json:"crawl_id"`
	Scheduled      bool      `json:"scheduled"`
	Start          time.Time `json:"start"`
	End            time.Time `json:"end"`
	TotalURLs      int       `json:"total_urls"`
	TotalIssues    int       `json:"total_issues"`
	CriticalIssues int       `json:"critical_issues"`
	AlertIssues    int       `json:"alert_issues"`
	WarningIssues  int       `json:"warning_issues"`
}

// WebhookService sends crawl notifications to a project's webhook URL.
type WebhookService struct {
	client *http.Client
}

func NewWebhookService() *WebhookService {
	return &WebhookService{
		client: &http.Client{Timeout: WebhookTimeout},
	}
}

// Notify posts a crawl summary to the project's webhook URL. It is a no-op
// when the project has no webhook URL configured, and it never fails the
// crawl: delivery errors are logged only.
func (s *WebhookService) Notify(p models.Project, c models.Crawl, scheduled bool) {
	if p.WebhookURL == "" {
		return
	}

	payload := CrawlNotification{
		ProjectID:      p.Id,
		URL:            p.URL,
		CrawlID:        c.Id,
		Scheduled:      scheduled,
		Start:          c.Start,
		End:            c.End,
		TotalURLs:      c.TotalURLs,
		TotalIssues:    c.TotalIssues,
		CriticalIssues: c.CriticalIssues,
		AlertIssues:    c.AlertIssues,
		WarningIssues:  c.WarningIssues,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Notify: error marshaling crawl notification for project %d: %v", p.Id, err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, p.WebhookURL, bytes.NewReader(body))
	if err != nil {
		log.Printf("Notify: error creating request for project %d: %v", p.Id, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("Notify: error sending crawl notification for project %d: %v", p.Id, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Printf("Notify: webhook for project %d returned status %d", p.Id, resp.StatusCode)
		return
	}

	log.Printf("Notify: sent crawl notification for project %d", p.Id)
}

// validateWebhookURL checks the project's webhook URL. An empty URL is valid
// and disables notifications, otherwise it must be an absolute http or https URL.
func validateWebhookURL(rawURL string) error {
	if rawURL == "" {
		return nil
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ErrWebhookURL
	}

	if parsedURL.Host == "" {
		return ErrWebhookURL
	}

	return nil
}
