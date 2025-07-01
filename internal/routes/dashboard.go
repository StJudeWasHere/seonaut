package routes

import (
	"net/http"
	"strconv"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/services"
)

type dashboardHandler struct {
	*services.Container
}

// indexHandler handles the dashboard of a project with all the needed data to render
// the charts. It expects a query parameter "pid" containing the project id.
func (h *dashboardHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
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

	data := struct {
		ProjectView       *models.ProjectView
		MediaChart        *models.Chart
		StatusChart       *models.Chart
		Crawls            []models.Crawl
		CanonicalCount    *models.CanonicalCount
		AltCount          *models.AltCount
		SchemeCount       *models.SchemeCount
		StatusCodeByDepth []models.StatusCodeByDepth
	}{
		ProjectView:       pv,
		MediaChart:        h.DashboardService.GetMediaCount(pv.Crawl.Id),
		StatusChart:       h.DashboardService.GetStatusCount(pv.Crawl.Id),
		Crawls:            h.CrawlerService.GetLastCrawls(pv.Project),
		CanonicalCount:    h.DashboardService.GetCanonicalCount(pv.Crawl.Id),
		AltCount:          h.DashboardService.GetImageAltCount(pv.Crawl.Id),
		SchemeCount:       h.DashboardService.GetSchemeCount(pv.Crawl.Id),
		StatusCodeByDepth: h.DashboardService.GetStatusCodeByDepth(pv.Crawl.Id),
	}

	pageView := &PageView{
		Data:      data,
		User:      *user,
		PageTitle: "PROJECT_DASHBOARD_PAGE_TITLE",
	}

	h.Renderer.RenderTemplate(w, "dashboard", pageView)
}
