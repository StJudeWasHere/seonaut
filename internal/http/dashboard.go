package http

import (
	"net/http"
	"strconv"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/services"
)

const (
	chartLimit = 4
)

type dashboardHandler struct {
	*services.Container
}

type ChartItem struct {
	Key   string
	Value int
}

type Chart []ChartItem

type DashboardView struct {
	ProjectView       *models.ProjectView
	MediaChart        Chart
	StatusChart       Chart
	Crawls            []models.Crawl
	CanonicalCount    *models.CanonicalCount
	AltCount          *models.AltCount
	SchemeCount       *models.SchemeCount
	StatusCodeByDepth []models.StatusCodeByDepth
}

// handleDashboard handles the dashboard of a project.
// It expects a query parameter "pid" containing the project ID.
func (h *dashboardHandler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)

		return
	}

	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	pv, err := h.ProjectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	if pv.Crawl.TotalURLs == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	data := DashboardView{
		ProjectView:       pv,
		MediaChart:        newChart(h.ReportService.GetMediaCount(pv.Crawl.Id)),
		StatusChart:       newChart(h.ReportService.GetStatusCount(pv.Crawl.Id)),
		Crawls:            h.CrawlerService.GetLastCrawls(pv.Project),
		CanonicalCount:    h.ReportService.GetCanonicalCount(pv.Crawl.Id),
		AltCount:          h.ReportService.GetImageAltCount(pv.Crawl.Id),
		SchemeCount:       h.ReportService.GetSchemeCount(pv.Crawl.Id),
		StatusCodeByDepth: h.ReportService.GetStatusCodeByDepth(pv.Crawl.Id),
	}

	pageView := &PageView{
		Data:      data,
		User:      *user,
		PageTitle: "PROJECT_DASHBOARD",
	}

	h.Renderer.RenderTemplate(w, "dashboard", pageView)
}

// Returns a Chart containing the keys and values from the CountList.
// It limits the slice to the chartLimit value.
func newChart(c *models.CountList) Chart {
	chart := Chart{}
	total := 0

	for _, i := range *c {
		total = total + i.Value
	}

	for _, i := range *c {
		ci := ChartItem{
			Key:   i.Key,
			Value: i.Value,
		}

		chart = append(chart, ci)
	}

	if len(chart) > chartLimit {
		chart[chartLimit-1].Key = "Other"
		for _, v := range chart[chartLimit:] {
			chart[chartLimit-1].Value += v.Value
		}

		chart = chart[:chartLimit]
	}

	return chart
}
