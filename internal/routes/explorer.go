package routes

import (
	"net/http"
	"strconv"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/services"
)

type explorerHandler struct {
	*services.Container
}

// indexHandler handles the URL explorer request.
// It performas a search of pagereports based on the "term" parameter. In case the "term" parameter
// is empty, it loads all the pagereports.
// It expects a query parameter "pid" containing the project id, the "p" parameter containing the current
// page in the paginator, and the "term" parameter used to perform the pagereport search.
func (h *explorerHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
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

	page, err := strconv.Atoi(r.URL.Query().Get("p"))
	if err != nil {
		page = 1
	}

	pv, err := h.ProjectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	term := r.URL.Query().Get("term")

	paginatorView, err := h.ReportService.GetPaginatedReports(pv.Crawl.Id, page, term)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	view := models.ExplorerView{
		ProjectView:   pv,
		Term:          term,
		PaginatorView: paginatorView,
	}

	v := &PageView{
		Data:      view,
		User:      *user,
		PageTitle: "EXPLORER_PAGE_TITLE",
	}

	h.Renderer.RenderTemplate(w, "explorer", v)
}
