package http

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/encoding"
	"github.com/stjudewashere/seonaut/internal/project"

	"github.com/turk/go-sitemap"
)

// serveExport handles the export request and renders the the export template
func (app *App) serveExport(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
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

	app.renderer.RenderTemplate(w, "export", &PageView{
		Data:      struct{ Project project.Project }{Project: pv.Project},
		User:      *user,
		PageTitle: "EXPORT_VIEW",
	})
}

func (app *App) serveDownloadCSV(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		log.Printf("serveDownloadCSV pid: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	pv, err := app.projectViewService.GetProjectView(pid, user.Id)
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
	prStream := app.reportService.GetPageReporsByIssueType(pv.Crawl.Id, eid)

	for p := range prStream {
		cw.Write(p)
	}
}

func (app *App) serveSitemap(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		log.Printf("serveSitemap pid: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	pv, err := app.projectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	w.Header().Add(
		"Content-Disposition",
		fmt.Sprint("attachment; filename=\""+pv.Project.Host+" "+time.Now().Format("2006-01-02")+" sitemap.xml\""))

	s := sitemap.NewSitemap(w, true)
	prStream := app.reportService.GetSitemapPageReports(pv.Crawl.Id)

	for v := range prStream {
		s.Add(v.URL, "")
	}

	s.Write()
}

// serveExportResources exports the resources of a specific project.
// The URL query parameter t specifys the type of resources to be exported.
func (app *App) serveExportResources(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	pv, err := app.projectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	t := r.URL.Query().Get("t")

	m := map[string]func(io.Writer, *crawler.Crawl){
		"internal":  app.exportService.ExportLinks,
		"external":  app.exportService.ExportExternalLinks,
		"images":    app.exportService.ExportImages,
		"scripts":   app.exportService.ExportScripts,
		"styles":    app.exportService.ExportStyles,
		"iframes":   app.exportService.ExportIframes,
		"audios":    app.exportService.ExportAudios,
		"videos":    app.exportService.ExportVideos,
		"hreflangs": app.exportService.ExportHreflangs,
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
