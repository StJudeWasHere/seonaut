package http

import (
	"net/http"
	"strconv"

	"github.com/stjudewashere/seonaut/internal/container"
	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/projectview"
)

type issueHandler struct {
	*container.Container
}

type IssuesGroupView struct {
	ProjectView *projectview.ProjectView
	IssueCount  *issue.IssueCount
}

type IssuesView struct {
	ProjectView   *projectview.ProjectView
	Eid           string
	PaginatorView models.PaginatorView
}

// handleIssues handles the issues view of a project.
// It expects a query parameter "pid" containing the project ID.
func (app *issueHandler) handleIssues(w http.ResponseWriter, r *http.Request) {
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

	ig := IssuesGroupView{
		ProjectView: pv,
		IssueCount:  app.IssueService.GetIssuesCount(pv.Crawl.Id),
	}

	v := &PageView{
		Data:      ig,
		User:      *user,
		PageTitle: "ISSUES_VIEW",
	}

	app.Renderer.RenderTemplate(w, "issues", v)
}

// handleIssuesView handles the view of project's specific issue type.
// It expects a query parameter "pid" containing the project ID and an "eid" parameter
// containing the issue type.
func (app *issueHandler) handleIssuesView(w http.ResponseWriter, r *http.Request) {
	user, ok := app.CookieSession.GetUser(r.Context())
	if !ok {
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

	pv, err := app.ProjectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	paginatorView, err := app.IssueService.GetPaginatedReportsByIssue(pv.Crawl.Id, page, eid)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	data := IssuesView{
		ProjectView:   pv,
		Eid:           eid,
		PaginatorView: paginatorView,
	}

	v := &PageView{
		Data:      data,
		User:      *user,
		PageTitle: "ISSUES_DETAIL",
	}

	app.Renderer.RenderTemplate(w, "issues_view", v)
}
