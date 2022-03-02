package http

import (
	"log"
	"net/http"
	"strconv"

	"github.com/stjudewashere/seonaut/internal/helper"
	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/project"
	"github.com/stjudewashere/seonaut/internal/user"
)

type IssuesGroupView struct {
	ProjectView *project.ProjectView
	MediaChart  helper.Chart
	StatusChart helper.Chart
	IssueCount  *issue.IssueCount
}

type IssuesView struct {
	ProjectView   *project.ProjectView
	Eid           string
	PaginatorView issue.PaginatorView
}

func (app *App) serveIssues(user *user.User, w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		log.Printf("serveIssues pid: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	pv, err := app.projectService.GetProjectView(pid, user.Id)
	if err != nil {
		log.Printf("serveIssues GetProjectView: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	issueCount := app.issueService.GetIssuesCount(pv.Crawl.Id)

	ig := IssuesGroupView{
		ProjectView: pv,
		MediaChart:  helper.NewChart(issueCount.MediaCount),
		StatusChart: helper.NewChart(issueCount.StatusCount),
		IssueCount:  issueCount,
	}

	v := &helper.PageView{
		Data:      ig,
		User:      *user,
		PageTitle: "ISSUES_VIEW",
	}

	app.renderer.RenderTemplate(w, "issues", v)
}

func (app *App) serveIssuesView(user *user.User, w http.ResponseWriter, r *http.Request) {
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

	pv, err := app.projectService.GetProjectView(pid, user.Id)
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

	v := &helper.PageView{
		Data:      view,
		User:      *user,
		PageTitle: "ISSUES_DETAIL",
	}

	app.renderer.RenderTemplate(w, "issues_view", v)
}
