package http

import (
	"net/http"
	"strconv"

	"github.com/stjudewashere/seonaut/internal/container"
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/projectview"
	"github.com/stjudewashere/seonaut/internal/report"
)

const (
	chartLimit = 4
)

type dashboardHandler struct {
	*container.Container
}

type ChartItem struct {
	Key   string
	Value int
}

type Chart []ChartItem

type DashboardView struct {
	ProjectView       *projectview.ProjectView
	MediaChart        Chart
	StatusChart       Chart
	Crawls            []models.Crawl
	CanonicalCount    *report.CanonicalCount
	AltCount          *report.AltCount
	SchemeCount       *report.SchemeCount
	StatusCodeByDepth []report.StatusCodeByDepth
}

// handleDashboard handles the dashboard of a project.
// It expects a query parameter "pid" containing the project ID.
func (app *dashboardHandler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := app.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)

		return
	}

	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	pv, err := app.ProjectViewService.GetProjectView(pid, user.Id)
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
		MediaChart:        newChart(app.ReportService.GetMediaCount(pv.Crawl.Id)),
		StatusChart:       newChart(app.ReportService.GetStatusCount(pv.Crawl.Id)),
		Crawls:            app.CrawlerService.GetLastCrawls(pv.Project),
		CanonicalCount:    app.ReportService.GetCanonicalCount(pv.Crawl.Id),
		AltCount:          app.ReportService.GetImageAltCount(pv.Crawl.Id),
		SchemeCount:       app.ReportService.GetSchemeCount(pv.Crawl.Id),
		StatusCodeByDepth: app.ReportService.GetStatusCodeByDepth(pv.Crawl.Id),
	}

	pageView := &PageView{
		Data:      data,
		User:      *user,
		PageTitle: "PROJECT_DASHBOARD",
	}

	app.Renderer.RenderTemplate(w, "dashboard", pageView)
}

// Returns a Chart containing the keys and values from the CountList.
// It limits the slice to the chartLimit value.
func newChart(c *report.CountList) Chart {
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
