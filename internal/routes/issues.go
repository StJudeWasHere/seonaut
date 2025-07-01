package routes

import (
	"net/http"
	"strconv"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/services"
)

type issueHandler struct {
	*services.Container
}

// indexHandler handles the issues view of a project.
// It expects a query parameter "pid" containing the project id.
func (h *issueHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
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

	ig := models.IssuesGroupView{
		ProjectView: pv,
		IssueCount:  h.IssueService.GetIssuesCount(pv.Crawl.Id),
	}

	v := &PageView{
		Data:      ig,
		User:      *user,
		PageTitle: "ISSUES_VIEW_PAGE_TITLE",
	}

	h.Renderer.RenderTemplate(w, "issues", v)
}

// viewHandler handles the view of the project's issues by an specific type.
// It expects a query parameter "pid" containing the project id and an "eid" parameter
// containing the issue type.
func (h *issueHandler) viewHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := h.CookieSession.GetUser(r.Context())
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

	pv, err := h.ProjectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	paginatorView, err := h.IssueService.GetPaginatedReportsByIssue(pv.Crawl.Id, page, eid)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := models.IssuesView{
		ProjectView:   pv,
		Eid:           eid,
		PaginatorView: paginatorView,
	}

	v := &PageView{
		Data:      data,
		User:      *user,
		PageTitle: "ISSUES_DETAIL_PAGE_TITLE",
	}

	h.Renderer.RenderTemplate(w, "issues_view", v)
}
