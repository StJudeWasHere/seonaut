package http

import (
	"log"
	"net/http"
	"strconv"

	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/projectview"
)

const (
	chartLimit = 4
)

type IssuesGroupView struct {
	ProjectView    *projectview.ProjectView
	MediaChart     Chart
	StatusChart    Chart
	IssueCount     *issue.IssueCount
	Crawls         []crawler.Crawl
	LinksCount     *issue.LinksCount
	CanonicalCount *issue.CanonicalCount
	AltCount       *issue.AltCount
	SchemeCount    *issue.SchemeCount
}

type IssuesView struct {
	ProjectView   *projectview.ProjectView
	Eid           string
	PaginatorView issue.PaginatorView
}

type ChartItem struct {
	Key   string
	Value int
}

type Chart []ChartItem

func (app *App) serveIssues(w http.ResponseWriter, r *http.Request) {
	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		log.Printf("serveIssues pid: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	pv, err := app.projectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		log.Printf("serveIssues GetProjectView: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if pv.Crawl.TotalURLs == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	issueCount := app.issueService.GetIssuesCount(pv.Crawl.Id)

	ig := IssuesGroupView{
		ProjectView: pv,
		MediaChart:  newChart(issueCount.MediaCount),
		StatusChart: newChart(issueCount.StatusCount),
		IssueCount:  issueCount,
		Crawls:      app.crawlerService.GetLastCrawls(pv.Project),
	}

	v := &PageView{
		Data:      ig,
		User:      *user,
		PageTitle: "ISSUES_VIEW",
	}

	app.renderer.RenderTemplate(w, "issues", v)
}

func (app *App) serveDashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		log.Printf("serveIssues pid: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	pv, err := app.projectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		log.Printf("serveIssues GetProjectView: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if pv.Crawl.TotalURLs == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	issueCount := app.issueService.GetIssuesCount(pv.Crawl.Id)

	ig := IssuesGroupView{
		ProjectView:    pv,
		MediaChart:     newChart(issueCount.MediaCount),
		StatusChart:    newChart(issueCount.StatusCount),
		IssueCount:     issueCount,
		Crawls:         app.crawlerService.GetLastCrawls(pv.Project),
		LinksCount:     app.issueService.GetLinksCount(pv.Crawl.Id),
		CanonicalCount: app.issueService.GetCanonicalCount(pv.Crawl.Id),
		AltCount:       app.issueService.GetImageAltCount(pv.Crawl.Id),
		SchemeCount:    app.issueService.GetSchemeCount(pv.Crawl.Id),
	}

	v := &PageView{
		Data:      ig,
		User:      *user,
		PageTitle: "PROJECT_DASHBOARD",
	}

	app.renderer.RenderTemplate(w, "dashboard", v)
}

func (app *App) serveIssuesView(w http.ResponseWriter, r *http.Request) {
	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	eid := r.URL.Query().Get("eid")
	if eid == "" {
		log.Println("serveIssuesView: eid parameter missing")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		log.Printf("serveIssuesView pid: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	page, err := strconv.Atoi(r.URL.Query().Get("p"))
	if err != nil {
		page = 1
	}

	pv, err := app.projectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		log.Printf("serveIssuesView GetProjectView: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	paginatorView, err := app.issueService.GetPaginatedReportsByIssue(pv.Crawl.Id, page, eid)
	if err != nil {
		log.Printf("serveIssuesView GetPaginatedReportsByIssue: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	view := IssuesView{
		ProjectView:   pv,
		Eid:           eid,
		PaginatorView: paginatorView,
	}

	v := &PageView{
		Data:      view,
		User:      *user,
		PageTitle: "ISSUES_DETAIL",
	}

	app.renderer.RenderTemplate(w, "issues_view", v)
}

func newChart(c issue.CountList) Chart {
	chart := Chart{}
	total := 0

	for _, i := range c {
		total = total + i.Value
	}

	for _, i := range c {
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
