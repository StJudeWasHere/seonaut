package routes

import (
	"net/http"

	"github.com/stjudewashere/seonaut/internal/services"
)

type supportHandler struct {
	*services.Container
}

// handleSupportSEOnaut handles the request for the "suppoert SEOnaut" page.
func (h *supportHandler) handleSupportSEOnaut(w http.ResponseWriter, r *http.Request) {
	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	pageView := &PageView{
		User:      *user,
		PageTitle: "SUPPORT_SEONAUT_VIEW_PAGE_TITLE",
	}

	h.Renderer.RenderTemplate(w, "support_seonaut", pageView)
}
