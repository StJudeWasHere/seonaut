package services_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/services"
)

func TestWebhookURLs(t *testing.T) {
	table := []struct {
		name      string
		webhook   string
		wantError bool
	}{
		{name: "No webhook", webhook: "", wantError: false},
		{name: "Valid https webhook", webhook: "https://hooks.example.com/seonaut", wantError: false},
		{name: "Valid http webhook", webhook: "http://localhost:3000/hook", wantError: false},
		{name: "Not supported scheme", webhook: "ftp://example.org/hook", wantError: true},
		{name: "Missing host", webhook: "https:///hook", wantError: true},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			p := &models.Project{URL: projectURL, UserAgent: userAgent, WebhookURL: tt.webhook}
			err := service.SaveProject(p, guid)
			if (err != nil) != tt.wantError {
				t.Errorf("SaveProject() want error %v got %v", tt.wantError, err)
			}
		})
	}
}

func TestNotifyPostsCrawlSummary(t *testing.T) {
	var received services.CrawlNotification

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", r.Header.Get("Content-Type"))
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		if err := json.Unmarshal(body, &received); err != nil {
			t.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := services.NewWebhookService()
	p := models.Project{Id: 3, URL: "https://example.com", WebhookURL: server.URL}
	crawl := models.Crawl{
		Id:             42,
		TotalURLs:      10,
		TotalIssues:    5,
		CriticalIssues: 1,
		AlertIssues:    2,
		WarningIssues:  2,
	}

	svc.Notify(p, crawl, true)

	if received.ProjectID != 3 {
		t.Errorf("project_id = %d, want 3", received.ProjectID)
	}

	if received.CrawlID != 42 {
		t.Errorf("crawl_id = %d, want 42", received.CrawlID)
	}

	if !received.Scheduled {
		t.Error("scheduled = false, want true")
	}

	if received.TotalURLs != 10 || received.TotalIssues != 5 {
		t.Errorf("totals = %+v, want 10 urls and 5 issues", received)
	}
}

func TestNotifySkipsEmptyURL(t *testing.T) {
	// Must not send anything and must not panic.
	services.NewWebhookService().Notify(
		models.Project{Id: 4, URL: "https://example.com"},
		models.Crawl{Id: 43},
		false,
	)
}
