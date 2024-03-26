package http

import (
	"log"
	"net/http"

	"github.com/stjudewashere/seonaut/internal/services"
)

type userHandler struct {
	*services.Container
}

// handleSignup handles the signup functionality for the application.
// It allows users to sign up by providing their email and password.
// Upon successful signup, the user is automatically signed in and redirected to the home page.
//
// The function handles both GET and POST HTTP methods.
// GET: Renders the signup form.
// POST: Processes the signup form data, performs signup, signs the user in, and redirects to the home page.
func (h *userHandler) handleSignup(w http.ResponseWriter, r *http.Request) {
	data := &struct {
		Email string
		Error bool
	}{}

	pageView := &PageView{
		PageTitle: "SIGNUP_VIEW",
		Data:      data,
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("serveSignup ParseForm: %v", err)
			http.Redirect(w, r, "/signup", http.StatusSeeOther)

			return
		}

		data.Email = r.FormValue("email")
		password := r.FormValue("password")

		u, err := h.UserService.SignUp(data.Email, password)
		if err != nil {
			log.Printf("serveSignup SignUp: %v", err)
			data.Error = true
			h.Renderer.RenderTemplate(w, "signup", pageView)

			return
		}

		h.CookieSession.SetSession(u.Id, w, r)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	h.Renderer.RenderTemplate(w, "signup", pageView)
}

// handleDeleteUser handles the HTTP GET and POST requests for the delete user account functionality.
//
// The function handles both GET and POST HTTP methods.
// GET: it renders the delete page with the appropriate data.
// POST: it sign's out the user and deletes the account including all its associated data.
func (h *userHandler) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)

		return
	}

	data := &struct {
		Crawling bool
	}{}

	pageView := &PageView{
		PageTitle: "DELETE_ACCOUNT_VIEW",
		User:      *user,
		Data:      data,
	}

	pv := h.ProjectViewService.GetProjectViews((user.Id))
	for _, v := range pv {
		if !v.Crawl.IssuesEnd.Valid || v.Project.Deleting {
			data.Crawling = true
			h.Renderer.RenderTemplate(w, "delete_account", pageView)
			return
		}
	}

	if r.Method == http.MethodPost {
		h.CookieSession.DestroySession(w, r)

		h.UserService.DeleteUser(user)

		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	h.Renderer.RenderTemplate(w, "delete_account", pageView)
}

// handleSignin handles the HTTP GET and POST requests for the sign-in functionality.
//
// The function handles both GET and POST HTTP methods.
// GET: it renders the sign-in page with the appropriate data.
// POST: it validates the user's credentials and creates a session if the sign-in is successful.
func (h *userHandler) handleSignin(w http.ResponseWriter, r *http.Request) {
	data := &struct {
		Email string
		Error bool
	}{}

	pageView := &PageView{
		PageTitle: "SIGNIN_VIEW",
		Data:      data,
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("serveSignin ParseForm: %v\n", err)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)

			return
		}

		data.Email = r.FormValue("email")
		password := r.FormValue("password")

		u, err := h.UserService.SignIn(data.Email, password)
		if err == nil {
			h.CookieSession.SetSession(u.Id, w, r)

			http.Redirect(w, r, "/", http.StatusSeeOther)

			return
		}

		log.Printf("serveSignin SignIn: %v\n", err)
		data.Error = true
	}

	h.Renderer.RenderTemplate(w, "signin", pageView)
}

// handleAccount handles the HTTP POST and GET requests for the account management functionality.
//
// The function handles both GET and POST HTTP methods.
// POST: it allows users to change their credentials by verifying the current password and
// updating the password with a new one.
// GET: it renders the account management form with the appropriate data.
func (h *userHandler) handleAccount(w http.ResponseWriter, r *http.Request) {
	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)

		return
	}

	data := &struct {
		Error        bool
		ErrorMessage string
	}{}

	pageView := &PageView{
		PageTitle: "ACCOUNT_VIEW",
		Data:      data,
		User:      *user,
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)

			return
		}

		password := r.FormValue("password")
		newPassword := r.FormValue("new_password")

		_, err = h.UserService.SignIn(user.Email, password)
		if err != nil {
			data.Error = true
			data.ErrorMessage = "Current password is not correct."

			h.Renderer.RenderTemplate(w, "account", pageView)

			return
		}

		err = h.UserService.UpdatePassword(user.Email, newPassword)
		if err != nil {
			data.Error = true
			data.ErrorMessage = "New password is not valid."

			h.Renderer.RenderTemplate(w, "account", pageView)

			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	h.Renderer.RenderTemplate(w, "account", pageView)
}

// handleSignout is a handler function that handles the user's signout request.
// It clears the session data related to authenticated user.
func (h *userHandler) handleSignout(w http.ResponseWriter, r *http.Request) {
	h.CookieSession.DestroySession(w, r)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
