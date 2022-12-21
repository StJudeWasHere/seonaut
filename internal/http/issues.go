package http

import (
	"net/http"
	"strconv"

	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/projectview"
)

type IssuesGroupView struct {
	ProjectView *projectview.ProjectView
	IssueCount  *issue.IssueCount
}

type IssuesView struct {
	ProjectView   *projectview.ProjectView
	Eid           string
	PaginatorView issue.PaginatorView
}

func (app *App) serveIssues(w http.ResponseWriter, r *http.Request) {
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

	ig := IssuesGroupView{
		ProjectView: pv,
		IssueCount:  app.issueService.GetIssuesCount(pv.Crawl.Id),
	}

	v := &PageView{
		Data:      ig,
		User:      *user,
		PageTitle: "ISSUES_VIEW",
	}

	app.renderer.RenderTemplate(w, "issues", v)
}

func (app *App) serveIssuesView(w http.ResponseWriter, r *http.Request) {
	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	eid := r.URL.Query().Get("eid")
	if eid == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	page, err := strconv.Atoi(r.URL.Query().Get("p"))
	if err != nil {
		page = 1
	}

	pv, err := app.projectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	paginatorView, err := app.issueService.GetPaginatedReportsByIssue(pv.Crawl.Id, page, eid)
	if err != nil {
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
