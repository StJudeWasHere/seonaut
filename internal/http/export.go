package http

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/stjudewashere/seonaut/internal/encoding"
	"github.com/stjudewashere/seonaut/internal/user"

	"github.com/turk/go-sitemap"
)

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

	pv, err := app.projectService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	eid := r.URL.Query().Get("eid")
	fileName := pv.Project.Host + " crawl " + time.Now().Format("2-15-2006")
	if eid != "" {
		fileName = fileName + "-" + eid
	}

	pageReports := app.reportService.GetPageReporsByIssueType(pv.Crawl.Id, eid)

	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.csv\"", fileName))

	cw := encoding.NewCSVWriter(w)
	for _, p := range pageReports {
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

	pv, err := app.projectService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	w.Header().Add(
		"Content-Disposition",
		fmt.Sprint("attachment; filename=\""+pv.Project.Host+" "+time.Now().Format("2-15-2006")+" sitemap.xml\""))

	s := sitemap.NewSitemap(w, true)
	p := app.reportService.GetSitemapPageReports(pv.Crawl.Id)
	for _, v := range p {
		s.Add(v.URL, "")
	}

	s.Write()
}
