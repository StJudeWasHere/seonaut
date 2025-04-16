package routes

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/services"
)

type replayHandler struct {
	*services.Container
}

// proxyHandler handles the HTTP request for the archive replay page. It loads the data from the
// WACZ archive and writes it to the ResponseWriter. The original HTML body of the resposne is rewritten
// so further requests to the links and relevant urls are routed through the proxyHandler.
func (h *replayHandler) proxyHandler(w http.ResponseWriter, r *http.Request) {
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

	pv, err := h.ProjectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	isArchived := h.Container.ArchiveService.ArchiveExists(&pv.Project)
	if !isArchived {
		http.Error(w, "Web archive does not exist", http.StatusNotFound)
		return
	}

	requestedURL := r.URL.Query().Get("url")
	if requestedURL == "" {
		http.Error(w, "URL parameter is missing", http.StatusInternalServerError)
		return
	}

	data := &struct {
		RequestedURL string
		ProjectView  *models.ProjectView
	}{
		ProjectView:  pv,
		RequestedURL: requestedURL,
	}

	// Avoid the browser cache
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	record, err := h.Container.ArchiveService.ReadArchiveRecord(&pv.Project, requestedURL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.Container.Renderer.RenderTemplate(w, "replay_not_archived", data)
		return
	}

	for key, values := range record.Headers {
		switch strings.ToLower(key) {
		case "set-cookie",
			"strict-transport-security",
			"content-security-policy",
			"access-control-allow-origin",
			"x-frame-options":
			continue // Skip unsafe headers
		default:
			for _, v := range values {
				w.Header().Add(key, v)
			}
		}
	}

	contentType := record.Headers.Get("Content-Type")
	isHTML := strings.HasPrefix(strings.ToLower(contentType), "text/html")

	if !isHTML {
		w.Write([]byte(record.Body))
		return
	}

	eb := new(bytes.Buffer)
	h.Container.Renderer.RenderTemplate(eb, "replay", data)

	rawBody := []byte(record.Body)
	rewrittenHTML, err := h.Container.ReplayService.RewriteHTML(rawBody, &pv.Project)
	if err != nil {
		http.Error(w, "Replay error", http.StatusInternalServerError)
		return
	}

	finalHTML, err := h.Container.ReplayService.InjectHTML(rewrittenHTML, eb.String())
	if err != nil {
		http.Error(w, "Replay error", http.StatusInternalServerError)
		return
	}

	w.Write(finalHTML)
}
