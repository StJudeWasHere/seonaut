package http

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/stjudewashere/seonaut/internal/container"
	"github.com/stjudewashere/seonaut/internal/models"
)

type projectHandler struct {
	*container.Container
}

// Handles the user homepage request and lists all the user's projects.
func (h *projectHandler) handleHome(w http.ResponseWriter, r *http.Request) {
	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	views := h.ProjectViewService.GetProjectViews(user.Id)

	var refresh bool
	for _, v := range views {
		if v.Crawl.Id > 0 && (!v.Crawl.IssuesEnd.Valid || v.Project.Deleting) {
			refresh = true
		}
	}

	v := &PageView{
		Data: struct {
			Projects []container.ProjectView
		}{Projects: views},
		User:      *user,
		PageTitle: "PROJECTS_VIEW",
		Refresh:   refresh,
	}

	h.Renderer.RenderTemplate(w, "home", v)
}

// handleProjectAdd handles the form for adding a new project.
func (h *projectHandler) handleProjectAdd(w http.ResponseWriter, r *http.Request) {
	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)

		return
	}

	data := &struct{ Error bool }{}

	pageView := &PageView{
		User:      *user,
		PageTitle: "ADD_PROJECT",
		Data:      data,
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("serveProjectAdd ParseForm: %v\n", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)

			return
		}

		u := r.FormValue("url")

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

		allowSubdomains, err := strconv.ParseBool(r.FormValue("allow_subdomains"))
		if err != nil {
			allowSubdomains = false
		}

		checkExternalLinks, err := strconv.ParseBool(r.FormValue("check_external_links"))
		if err != nil {
			checkExternalLinks = false
		}

		basicAuth, err := strconv.ParseBool(r.FormValue("basic_auth"))
		if err != nil {
			basicAuth = false
		}

		parsedURL, err := url.ParseRequestURI(strings.TrimSpace(u))
		if err != nil {
			data.Error = true
			h.Renderer.RenderTemplate(w, "project_add", pageView)

			return
		}

		project := &models.Project{
			URL:                parsedURL.String(),
			IgnoreRobotsTxt:    ignoreRobotsTxt,
			FollowNofollow:     followNofollow,
			IncludeNoindex:     includeNoindex,
			CrawlSitemap:       crawlSitemap,
			AllowSubdomains:    allowSubdomains,
			BasicAuth:          basicAuth,
			CheckExternalLinks: checkExternalLinks,
		}

		err = h.ProjectService.SaveProject(project, user.Id)
		if err != nil {
			data.Error = true
			h.Renderer.RenderTemplate(w, "project_add", pageView)

			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	h.Renderer.RenderTemplate(w, "project_add", pageView)
}

// handleDeleteProject handles the deletion of a project.
// It expects a query parameter "pid" containing the project ID to be deleted.
func (h *projectHandler) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)

		return
	}

	p, err := h.ProjectService.FindProject(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	h.ProjectService.DeleteProject(&p)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleProjectEdit handles the edition of a project.
// It expects a query parameter "pid" containing the project ID to be edited.
func (h *projectHandler) handleProjectEdit(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)

		return
	}

	p, err := h.ProjectService.FindProject(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	data := &struct {
		Project models.Project
		Error   bool
	}{
		Project: p,
	}

	pageView := &PageView{
		User:      *user,
		PageTitle: "EDIT_PROJECT",
		Data:      data,
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("serveProjectEdit ParseForm: %v\n", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		p.IgnoreRobotsTxt, err = strconv.ParseBool(r.FormValue("ignore_robotstxt"))
		if err != nil {
			p.IgnoreRobotsTxt = false
		}

		p.FollowNofollow, err = strconv.ParseBool(r.FormValue("follow_nofollow"))
		if err != nil {
			p.FollowNofollow = false
		}

		p.IncludeNoindex, err = strconv.ParseBool(r.FormValue("include_noindex"))
		if err != nil {
			p.IncludeNoindex = false
		}

		p.CrawlSitemap, err = strconv.ParseBool(r.FormValue("crawl_sitemap"))
		if err != nil {
			p.CrawlSitemap = false
		}

		p.AllowSubdomains, err = strconv.ParseBool(r.FormValue("allow_subdomains"))
		if err != nil {
			p.AllowSubdomains = false
		}

		p.CheckExternalLinks, err = strconv.ParseBool(r.FormValue("check_external_links"))
		if err != nil {
			p.CheckExternalLinks = false
		}

		p.BasicAuth, err = strconv.ParseBool(r.FormValue("basic_auth"))
		if err != nil {
			p.BasicAuth = false
		}

		err = h.ProjectService.UpdateProject(&p)
		if err != nil {
			data.Error = true
			h.Renderer.RenderTemplate(w, "project_edit", pageView)

			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	h.Renderer.RenderTemplate(w, "project_edit", pageView)
}
