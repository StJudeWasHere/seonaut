package http

import (
	"net/http"
	"strconv"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/projectview"
)

type ExplorerView struct {
	ProjectView   *projectview.ProjectView
	Term          string
	PaginatorView models.PaginatorView
}

// Handles the URL explorer request.
func (app *App) serveExplorer(w http.ResponseWriter, r *http.Request) {
	// Get user from the request's context
	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	// Get the project id
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Get the page number and set page number to 1 if the parameter is not set
	page, err := strconv.Atoi(r.URL.Query().Get("p"))
	if err != nil {
		page = 1
	}

	// Get the project view
	pv, err := app.projectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	term := r.URL.Query().Get("term")

	// Get the paginated reports
	paginatorView, err := app.reportService.GetPaginatedReports(pv.Crawl.Id, page, term)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	view := ExplorerView{
		ProjectView:   pv,
		Term:          term,
		PaginatorView: paginatorView,
	}

	v := &PageView{
		Data:      view,
		User:      *user,
		PageTitle: "EXPLORER",
	}

	app.renderer.RenderTemplate(w, "explorer", v)
}
