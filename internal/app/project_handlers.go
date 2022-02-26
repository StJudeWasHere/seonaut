package app

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mnlg/lenkrr/internal/user"
)

type ProjectView struct {
	Project Project
	Crawl   Crawl
}

type Project struct {
	Id              int
	URL             string
	Host            string
	IgnoreRobotsTxt bool
	UseJS           bool
	Created         time.Time
}

func (app *App) serveHome(user *user.User, w http.ResponseWriter, r *http.Request) {
	var refresh bool
	var views []ProjectView
	projects := app.datastore.findProjectsByUser(user.Id)

	for _, p := range projects {
		c := app.datastore.getLastCrawl(&p)
		pv := ProjectView{
			Project: p,
			Crawl:   c,
		}
		views = append(views, pv)

		if c.IssuesEnd.Valid == false {
			refresh = true
		}
	}

	v := &PageView{
		Data: struct {
			Projects    []ProjectView
			MaxProjects int
		}{Projects: views, MaxProjects: user.GetMaxAllowedProjects()},
		User:      *user,
		PageTitle: "PROJECTS_VIEW",
		Refresh:   refresh,
	}

	app.renderer.renderTemplate(w, "home", v)
}

func (app *App) serveProjectAdd(user *user.User, w http.ResponseWriter, r *http.Request) {
	projects := app.datastore.findProjectsByUser(user.Id)
	if len(projects) >= user.GetMaxAllowedProjects() {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			log.Println(err)
		}

		url := r.FormValue("url")

		ignoreRobotsTxt, err := strconv.ParseBool(r.FormValue("ignore_robotstxt"))
		if err != nil {
			ignoreRobotsTxt = false
		}

		useJavascript, err := strconv.ParseBool(r.FormValue("use_javascript"))
		if err != nil {
			useJavascript = false
		}

		if user.Advanced == false {
			useJavascript = false
		}

		app.datastore.saveProject(url, ignoreRobotsTxt, useJavascript, user.Id)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	v := &PageView{
		User:      *user,
		PageTitle: "ADD_PROJECT",
	}

	app.renderer.renderTemplate(w, "project_add", v)
}
