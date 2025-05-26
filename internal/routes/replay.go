package routes

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
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
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	data := &struct {
		RequestedURL string
		ProjectView  *models.ProjectView
	}{
		ProjectView:  pv,
		RequestedURL: requestedURL,
	}

	record, err := h.Container.ArchiveService.ReadArchiveRecord(&pv.Project, requestedURL)
	if err != nil {
		http.Error(w, "The requested URL is not archived.", http.StatusNotFound)
		return
	}

	// Send the headers in the archived response
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

	// Avoid the browser cache
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	parsedRequestedURL, err := url.Parse(requestedURL)
	if err != nil {
		w.Write([]byte(record.Body))
		return
	}

	// rewriteFunc is a function that rewrites URLs so they are proxied
	// through the replay proxy URL. This function is passed as a parameter
	// to the replay service rewrite methods and is used to rewrite the URLS
	// in the HTML and CSS files.
	rewriteFunc := func(urlStr string) string {
		if u, err := url.Parse(urlStr); err == nil {
			if !u.IsAbs() {
				u = parsedRequestedURL.ResolveReference(u)
			}

			if u.Scheme != "http" && u.Scheme != "https" {
				return urlStr
			}

			return fmt.Sprintf("/replay?pid=%d&url=%s", pv.Project.Id, u.String())
		}

		return urlStr
	}

	// If the requested URL is a CSS the URLs are rewritten and the resulting
	// content is sent to the client.
	contentType := record.Headers.Get("Content-Type")
	if strings.HasPrefix(strings.ToLower(contentType), "text/css") {
		rewrittenCSS := h.ReplayService.RewriteCSS(string(record.Body), rewriteFunc)

		w.Write([]byte(rewrittenCSS))
		return
	}

	// If the requested URL is not HTML it is sent as it is in the archive
	// as it may be an image, a JS file or any other resource.
	isHTML := strings.HasPrefix(strings.ToLower(contentType), "text/html")
	if !isHTML {
		w.Write([]byte(record.Body))
		return
	}

	// In case it is an HTML the links are rewritten so they go through the proxy.
	// Then the replay banner and replay scripts are rendered and injected into the HTML.
	rawBody := []byte(record.Body)
	rewrittenHTML, err := h.Container.ReplayService.RewriteHTML(rawBody, rewriteFunc)
	if err != nil {
		http.Error(w, "Replay error", http.StatusInternalServerError)
		return
	}

	eb := new(bytes.Buffer)
	h.Container.Renderer.RenderTemplate(eb, "replay_banner", data)

	es := new(bytes.Buffer)
	h.Container.Renderer.RenderTemplate(es, "replay_scripts", data)

	finalHTML, err := h.Container.ReplayService.InjectHTML(rewrittenHTML, es.String(), eb.String())
	if err != nil {
		http.Error(w, "Replay error", http.StatusInternalServerError)
		return
	}

	w.Write(finalHTML)
}
