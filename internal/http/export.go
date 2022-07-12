package http

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/stjudewashere/seonaut/internal/encoding"
	"github.com/stjudewashere/seonaut/internal/helper"
	"github.com/stjudewashere/seonaut/internal/project"
	"github.com/stjudewashere/seonaut/internal/user"

	"github.com/turk/go-sitemap"
)

func (app *App) serveExport(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		log.Printf("serveDownloadCSV pid: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	c := r.Context().Value("user")
	user, ok := c.(*user.User)
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	pv, err := app.projectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	type data struct{ Project project.Project }

	v := &helper.PageView{
		Data:      data{Project: pv.Project},
		User:      *user,
		PageTitle: "EXPORT_VIEW",
	}

	app.renderer.RenderTemplate(w, "export", v)
}

func (app *App) serveDownloadCSV(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		log.Printf("serveDownloadCSV pid: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	c := r.Context().Value("user")
	user, ok := c.(*user.User)
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
	fileName := pv.Project.Host + " crawl " + time.Now().Format("2-15-2006")
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

	c := r.Context().Value("user")
	user, ok := c.(*user.User)
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
		fmt.Sprint("attachment; filename=\""+pv.Project.Host+" "+time.Now().Format("2-15-2006")+" sitemap.xml\""))

	s := sitemap.NewSitemap(w, true)
	prStream := app.reportService.GetSitemapPageReports(pv.Crawl.Id)

	for v := range prStream {
		s.Add(v.URL, "")
	}

	s.Write()
}

func (app *App) serveDownload(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	t := r.URL.Query().Get("t")

	c := r.Context().Value("user")
	user, ok := c.(*user.User)
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	pv, err := app.projectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	fileName := pv.Project.Host + " " + t + " " + time.Now().Format("2-15-2006")

	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.csv\"", fileName))

	switch t {
	case "internal":
		app.exportService.ExportLinks(w, &pv.Crawl)
	case "external":
		app.exportService.ExportExternalLinks(w, &pv.Crawl)
	case "images":
		app.exportService.ExportImages(w, &pv.Crawl)
	case "scripts":
		app.exportService.ExportScripts(w, &pv.Crawl)
	case "styles":
		app.exportService.ExportStyles(w, &pv.Crawl)
	case "iframes":
		app.exportService.ExportIframes(w, &pv.Crawl)
	}
}
