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

type DashboardView struct {
	ProjectView       *models.ProjectView
	MediaChart        *models.Chart
	StatusChart       *models.Chart
	Crawls            []models.Crawl
	CanonicalCount    *models.CanonicalCount
	AltCount          *models.AltCount
	SchemeCount       *models.SchemeCount
	StatusCodeByDepth []models.StatusCodeByDepth
}

// handleDashboard handles the dashboard of a project.
// It expects a query parameter "pid" containing the project ID.
func (h *dashboardHandler) handleDashboard(w http.ResponseWriter, r *http.Request) {
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

	data := DashboardView{
		ProjectView:       pv,
		MediaChart:        h.ReportService.GetMediaCount(pv.Crawl.Id),
		StatusChart:       h.ReportService.GetStatusCount(pv.Crawl.Id),
		Crawls:            h.CrawlerService.GetLastCrawls(pv.Project),
		CanonicalCount:    h.ReportService.GetCanonicalCount(pv.Crawl.Id),
		AltCount:          h.ReportService.GetImageAltCount(pv.Crawl.Id),
		SchemeCount:       h.ReportService.GetSchemeCount(pv.Crawl.Id),
		StatusCodeByDepth: h.ReportService.GetStatusCodeByDepth(pv.Crawl.Id),
	}

	pageView := &PageView{
		Data:      data,
		User:      *user,
		PageTitle: "PROJECT_DASHBOARD",
	}

	h.Renderer.RenderTemplate(w, "dashboard", pageView)
}
