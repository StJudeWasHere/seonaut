package routes

import (
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/services"
)

type archiveHandler struct {
	*services.Container
}

// archiveHandler the HTTP request for the archive page. It loads the data from the
// archive and displays the source code of the crawler's response for a specific resource.
func (h *archiveHandler) archiveHandler(w http.ResponseWriter, r *http.Request) {
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

	rid, err := strconv.Atoi(r.URL.Query().Get("rid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	eid := r.URL.Query().Get("eid")
	ep := r.URL.Query().Get("ep")
	if eid == "" && ep == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	pv, err := h.ProjectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	isArchived := h.Container.ArchiveService.ArchiveExists(&pv.Project)
	if !isArchived {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	pageReportView := h.ReportService.GetPageReport(rid, pv.Crawl.Id, "default", 1)

	record, err := h.Container.ArchiveService.ReadArchiveRecord(&pv.Project, pageReportView.PageReport.URL)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	isText := strings.HasPrefix(pageReportView.PageReport.MediaType, "text/")

	data := &struct {
		PageReportView *models.PageReportView
		ProjectView    *models.ProjectView
		Eid            string
		Ep             string
		ArchiveRecord  *models.ArchiveRecord
		IsText         bool
	}{
		ProjectView:    pv,
		PageReportView: pageReportView,
		Eid:            eid,
		Ep:             ep,
		ArchiveRecord:  record,
		IsText:         isText,
	}

	pageView := &PageView{
		Data:      data,
		User:      *user,
		PageTitle: "ARCHIVE_VIEW_PAGE_TITLE",
	}

	h.Renderer.RenderTemplate(w, "archive", pageView)
}

// downloadHandler allows to download an archived resource. It loads the data from the
// archive and and sets the headers to force the download.
func (h *archiveHandler) downloadHandler(w http.ResponseWriter, r *http.Request) {
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

	rid, err := strconv.Atoi(r.URL.Query().Get("rid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	pv, err := h.ProjectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	isArchived := h.Container.ArchiveService.ArchiveExists(&pv.Project)
	if !isArchived {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	pageReportView := h.ReportService.GetPageReport(rid, pv.Crawl.Id, "default", 1)

	record, err := h.Container.ArchiveService.ReadArchiveRecord(&pv.Project, pageReportView.PageReport.URL)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	pathPart := strings.Trim(pageReportView.PageReport.ParsedURL.Path, "/")
	fileName := path.Base(pathPart)

	if fileName == "/" || fileName == "." {
		fileName = "index.html"
	}

	if strings.HasPrefix(pageReportView.PageReport.MediaType, "text/html") && !strings.HasSuffix(strings.ToLower(fileName), ".html") {
		fileName += ".html"
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Expires", "0")
	w.Header().Set("Cache-Control", "must-revalidate")
	w.Header().Set("Pragma", "public")

	w.Write([]byte(record.Body))
}
