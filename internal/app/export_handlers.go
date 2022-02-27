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
	qcid, ok := r.URL.Query()["cid"]
	if !ok || len(qcid) < 1 {
		log.Println("serveDownloadCSV: cid parameter missing")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	cid, err := strconv.Atoi(qcid[0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	u, err := app.datastore.findCrawlUserId(cid)
	if err != nil || u.Id != user.Id {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	crawl := app.datastore.findCrawlById(cid)

	project, err := app.projectService.FindProject(crawl.ProjectId, user.Id)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	var pageReports []report.PageReport

	eid := r.URL.Query()["eid"]
	fileName := project.Host + " crawl " + time.Now().Format("2-15-2006")

	if len(eid) > 0 && eid[0] != "" {
		fileName = fileName + "-" + eid[0]
		pageReports = app.datastore.FindAllPageReportsByCrawlIdAndErrorType(cid, eid[0])
	} else {
		pageReports = app.datastore.FindAllPageReportsByCrawlId(cid)
	}

	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.csv\"", fileName))

	cw := encoding.NewCSVWriter(w)
	for _, p := range pageReports {
		cw.Write(p)
	}
}

func (app *App) serveSitemap(user *user.User, w http.ResponseWriter, r *http.Request) {
	qcid, ok := r.URL.Query()["cid"]
	if !ok || len(qcid) < 1 {
		log.Println("serveSitemap: cid parameter missings")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	cid, err := strconv.Atoi(qcid[0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	u, err := app.datastore.findCrawlUserId(cid)
	if err != nil || u.Id != user.Id {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	crawl := app.datastore.findCrawlById(cid)
	project, err := app.projectService.FindProject(crawl.ProjectId, user.Id)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	w.Header().Add(
		"Content-Disposition",
		fmt.Sprint("attachment; filename=\""+project.Host+" "+time.Now().Format("2-15-2006")+" sitemap.xml\""))

	s := sitemap.NewSitemap(w, true)
	p := app.datastore.findSitemapPageReports(cid)
	for _, v := range p {
		s.Add(v.URL, "")
	}

	s.Write()
}
