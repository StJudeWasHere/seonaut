package http

import (
	"net/http"
	"strconv"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/projectview"
	"github.com/stjudewashere/seonaut/internal/report"
)

const (
	chartLimit = 4
)

type ChartItem struct {
	Key   string
	Value int
}

type Chart []ChartItem

type DashboardView struct {
	ProjectView    *projectview.ProjectView
	MediaChart     Chart
	StatusChart    Chart
	Crawls         []models.Crawl
	CanonicalCount *report.CanonicalCount
	AltCount       *report.AltCount
	SchemeCount    *report.SchemeCount
}

func (app *App) serveDashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	pv, err := app.projectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if pv.Crawl.TotalURLs == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	ig := DashboardView{
		ProjectView:    pv,
		MediaChart:     newChart(app.reportService.GetMediaCount(pv.Crawl.Id)),
		StatusChart:    newChart(app.reportService.GetStatusCount(pv.Crawl.Id)),
		Crawls:         app.crawlerService.GetLastCrawls(pv.Project),
		CanonicalCount: app.reportService.GetCanonicalCount(pv.Crawl.Id),
		AltCount:       app.reportService.GetImageAltCount(pv.Crawl.Id),
		SchemeCount:    app.reportService.GetSchemeCount(pv.Crawl.Id),
	}

	v := &PageView{
		Data:      ig,
		User:      *user,
		PageTitle: "PROJECT_DASHBOARD",
	}

	app.renderer.RenderTemplate(w, "dashboard", v)
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
