package app

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mnlg/lenkrr/internal/encoding"
	"github.com/mnlg/lenkrr/internal/report"
	"github.com/mnlg/lenkrr/internal/user"

	"github.com/turk/go-sitemap"
)

func (app *App) serveDownloadCSV(user *user.User, w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	pv, err := app.projectService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	var pageReports []report.PageReport

	eid := r.URL.Query().Get("eid")
	fileName := pv.Project.Host + " crawl " + time.Now().Format("2-15-2006")

	if eid != "" {
		fileName = fileName + "-" + eid
		pageReports = app.datastore.FindAllPageReportsByCrawlIdAndErrorType(pv.Crawl.Id, eid)
	} else {
		pageReports = app.datastore.FindAllPageReportsByCrawlId(pv.Crawl.Id)
	}

	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.csv\"", fileName))

	cw := encoding.NewCSVWriter(w)
	for _, p := range pageReports {
		cw.Write(p)
	}
}

func (app *App) serveSitemap(user *user.User, w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	pv, err := app.projectService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	w.Header().Add(
		"Content-Disposition",
		fmt.Sprint("attachment; filename=\""+pv.Project.Host+" "+time.Now().Format("2-15-2006")+" sitemap.xml\""))

	s := sitemap.NewSitemap(w, true)
	p := app.datastore.findSitemapPageReports(pv.Crawl.Id)
	for _, v := range p {
		s.Add(v.URL, "")
	}

	s.Write()
}
