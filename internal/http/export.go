package http

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/stjudewashere/seonaut/internal/container"
	"github.com/stjudewashere/seonaut/internal/encoding"
	"github.com/stjudewashere/seonaut/internal/models"

	"github.com/turk/go-sitemap"
)

type exportHandler struct {
	*container.Container
}

// handleExport handles the export request and renders the the export template.
func (app *exportHandler) handleExport(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	user, ok := app.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
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

	app.Renderer.RenderTemplate(w, "export", &PageView{
		Data:      struct{ Project models.Project }{Project: pv.Project},
		User:      *user,
		PageTitle: "EXPORT_VIEW",
	})
}

// handleDownloadCSV exports the pagereports of a specific project as a CSV file by issue type.
func (app *exportHandler) handleDownloadCSV(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	user, ok := app.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)

		return
	}

	pv, err := app.ProjectViewService.GetProjectView(pid, user.Id)
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

	cw := encoding.NewCSVWriter(w)
	prStream := app.ReportService.GetPageReporsByIssueType(pv.Crawl.Id, eid)

	for p := range prStream {
		cw.Write(p)
	}
}

// handleSitemap exports the crawled urls of a specific project as a sitemap.xml file.
func (app *exportHandler) handleSitemap(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	user, ok := app.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)

		return
	}

	pv, err := app.ProjectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	w.Header().Add(
		"Content-Disposition",
		fmt.Sprint("attachment; filename=\""+pv.Project.Host+" "+time.Now().Format("2006-01-02")+" sitemap.xml\""))

	s := sitemap.NewSitemap(w, true)
	prStream := app.ReportService.GetSitemapPageReports(pv.Crawl.Id)

	for v := range prStream {
		s.Add(v.URL, "")
	}

	s.Write()
}

// handleExportResources exports the resources of a specific project.
// The URL query parameter t specifys the type of resources to be exported.
func (app *exportHandler) handleExportResources(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	user, ok := app.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)

		return
	}

	pv, err := app.ProjectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	t := r.URL.Query().Get("t")

	m := map[string]func(io.Writer, *models.Crawl){
		"internal":  app.ExportService.ExportLinks,
		"external":  app.ExportService.ExportExternalLinks,
		"images":    app.ExportService.ExportImages,
		"scripts":   app.ExportService.ExportScripts,
		"styles":    app.ExportService.ExportStyles,
		"iframes":   app.ExportService.ExportIframes,
		"audios":    app.ExportService.ExportAudios,
		"videos":    app.ExportService.ExportVideos,
		"hreflangs": app.ExportService.ExportHreflangs,
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
