package http

import (
	"log"
	"net/http"
	"strconv"

	"github.com/stjudewashere/seonaut/internal/helper"
	"github.com/stjudewashere/seonaut/internal/projectview"
	"github.com/stjudewashere/seonaut/internal/user"
)

func (app *App) serveHome(w http.ResponseWriter, r *http.Request) {
	c := r.Context().Value("user")
	user, ok := c.(*user.User)
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	views := app.projectViewService.GetProjectViews(user.Id)

	var refresh bool
	for _, v := range views {
		if v.Crawl.IssuesEnd.Valid == false {
			refresh = true
		}
	}

	v := &helper.PageView{
		Data: struct {
			Projects []projectview.ProjectView
		}{Projects: views},
		User:      *user,
		PageTitle: "PROJECTS_VIEW",
		Refresh:   refresh,
	}

	app.renderer.RenderTemplate(w, "home", v)
}

func (app *App) serveProjectAdd(w http.ResponseWriter, r *http.Request) {
	c := r.Context().Value("user")
	user, ok := c.(*user.User)
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	var e bool

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("serveProjectAdd ParseForm: %v\n", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		url := r.FormValue("url")

		ignoreRobotsTxt, err := strconv.ParseBool(r.FormValue("ignore_robotstxt"))
		if err != nil {
			ignoreRobotsTxt = false
		}

		followNofollow, err := strconv.ParseBool(r.FormValue("follow_nofollow"))
		if err != nil {
			followNofollow = false
		}

		includeNoindex, err := strconv.ParseBool(r.FormValue("include_noindex"))
		if err != nil {
			includeNoindex = false
		}

		crawlSitemap, err := strconv.ParseBool(r.FormValue("crawl_sitemap"))
		if err != nil {
			crawlSitemap = false
		}

		err = app.projectService.SaveProject(url, ignoreRobotsTxt, followNofollow, includeNoindex, crawlSitemap, user.Id)

		if err == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		e = true

	}

	v := &helper.PageView{
		User:      *user,
		PageTitle: "ADD_PROJECT",
		Data:      struct{ Error bool }{Error: e},
	}

	app.renderer.RenderTemplate(w, "project_add", v)
}
