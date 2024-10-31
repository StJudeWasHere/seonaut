package routes

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/services"

	"github.com/turk/go-sitemap"
)

type exportHandler struct {
	*services.Container
}

// indexHandler handles the export request and renders the the export template.
// It expects a "pid" query parameter with the project's id.
func (h *exportHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
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

	h.Renderer.RenderTemplate(w, "export", &PageView{
		Data:      struct{ Project models.Project }{Project: pv.Project},
		User:      *user,
		PageTitle: "EXPORT_VIEW",
	})
}

// csvHandler exports the pagereports of a specific project as a CSV file by issue type.
// It expects a "pid" query parameter with the project's id. If the "eid" query parameter
// is set, it exports the pagereports with an specific issue type.
func (h *exportHandler) csvHandler(w http.ResponseWriter, r *http.Request) {
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

	eid := r.URL.Query().Get("eid")
	fileName := pv.Project.Host + " crawl " + time.Now().Format("2006-01-02")
	if eid != "" {
		fileName = fileName + "-" + eid
	}

	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.csv\"", fileName))

	cw := services.NewCSVWriter(w)
	prStream := h.ReportService.GetPageReporsByIssueType(pv.Crawl.Id, eid)

	for p := range prStream {
		cw.Write(p)
	}
}

// sitemapHandler exports the crawled urls of a specific project as a sitemap.xml file.
// It expects a "pid" query parameter with the project's id.
func (h *exportHandler) sitemapHandler(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Add(
		"Content-Disposition",
		fmt.Sprint("attachment; filename=\""+pv.Project.Host+" "+time.Now().Format("2006-01-02")+" sitemap.xml\""))

	s := sitemap.NewSitemap(w, true)
	prStream := h.ReportService.GetSitemapPageReports(pv.Crawl.Id)

	for v := range prStream {
		s.Add(v.URL, "")
	}

	s.Write()
}

// resourcesHandler exports the resources of a specific project.
// It expects a "pid" query parameter with the project's id as well as a query
// parameter "t" specifys the type of resources to be exported.
func (h *exportHandler) resourcesHandler(w http.ResponseWriter, r *http.Request) {
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

	t := r.URL.Query().Get("t")

	m := map[string]func(io.Writer, *models.Crawl){
		"internal":  h.ExportService.ExportLinks,
		"external":  h.ExportService.ExportExternalLinks,
		"images":    h.ExportService.ExportImages,
		"scripts":   h.ExportService.ExportScripts,
		"styles":    h.ExportService.ExportStyles,
		"iframes":   h.ExportService.ExportIframes,
		"audios":    h.ExportService.ExportAudios,
		"videos":    h.ExportService.ExportVideos,
		"hreflangs": h.ExportService.ExportHreflangs,
	}

	e, ok := m[t]
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	fileName := pv.Project.Host + " " + t + " " + time.Now().Format("2006-01-02")
	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.csv\"", fileName))
	e(w, &pv.Crawl)
}
