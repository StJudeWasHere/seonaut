package routes

import (
	"log"
	"net/http"

	"github.com/stjudewashere/seonaut/internal/services"
)

type userHandler struct {
	*services.Container
}

// signupGetHandler handles the GET signup request ehich displays the sign up form.
// It allows users to sign up by providing their email and password.
func (h *userHandler) signupGetHandler(w http.ResponseWriter, r *http.Request) {
	data := &struct {
		Email string
		Error bool
	}{}

	pageView := &PageView{
		PageTitle: "SIGNUP_VIEW",
		Data:      data,
	}

	h.Renderer.RenderTemplate(w, "signup", pageView)
}

// signupPostHandler handles the POST request of the signup form.
// Upon successful signup, the user is automatically signed in and redirected to the home page.
func (h *userHandler) signupPostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Printf("serveSignup ParseForm: %v", err)
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	u, err := h.UserService.SignUp(email, password)
	if err == nil {
		h.CookieSession.SetSession(u, w, r)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	log.Printf("serveSignup SignUp: %v", err)

	data := &struct {
		Email string
		Error bool
	}{
		Email: email,
		Error: true,
	}

	pageView := &PageView{
		PageTitle: "SIGNUP_VIEW",
		Data:      data,
	}

	h.Renderer.RenderTemplate(w, "signup", pageView)
}

// deleteGetHandler handles the HTTP GET request for the delete user account functionality.
// it renders the delete page with the appropriate data.
func (h *userHandler) deleteGetHandler(w http.ResponseWriter, r *http.Request) {
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
		if v.Crawl.Crawling || v.Project.Deleting {
			data.Crawling = true
			break
		}
	}

	h.Renderer.RenderTemplate(w, "delete_account", pageView)
}

// deletePostHandler handles the POST request to delete an account.
// After deleting the user and all its associated data it destroys the session
// and redirects home.
func (h *userHandler) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	h.UserService.DeleteUser(user)
	h.CookieSession.DestroySession(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// signinGetHandler handles the HTTP GET request for the sign-in functionality.
// It renders the sign-in page with the appropriate data.
func (h *userHandler) signinGetHandler(w http.ResponseWriter, r *http.Request) {
	pageView := &PageView{
		PageTitle: "SIGNIN_VIEW",
		Data: &struct {
			Email string
			Error bool
		}{},
	}

	h.Renderer.RenderTemplate(w, "signin", pageView)
}

// signinGetHandler handles the HTTP POST request for the sign-in functionality.
// It validates the user's credentials and creates a session if the sign-in is successful.
func (h *userHandler) signinPostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Printf("serveSignin ParseForm: %v\n", err)
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	u, err := h.UserService.SignIn(email, password)
	if err != nil {
		log.Printf("serveSignin SignIn: %v\n", err)

		pageView := &PageView{
			PageTitle: "SIGNIN_VIEW",
			Data: &struct {
				Email string
				Error bool
			}{
				Email: email,
				Error: true,
			},
		}

		h.Renderer.RenderTemplate(w, "signin", pageView)
		return
	}

	h.CookieSession.SetSession(u, w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// editGetHandler handles the HTTP GET request for the account management functionality.
// It renders the account management form with the appropriate data.
func (h *userHandler) editGetHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	pageView := &PageView{
		PageTitle: "ACCOUNT_VIEW",
		User:      *user,
		Data: &struct {
			Error        bool
			ErrorMessage string
		}{},
	}

	h.Renderer.RenderTemplate(w, "account", pageView)
}

// editPostHandler handles the HTTP POST request for the account management functionality.
// It allows users to change their credentials by verifying the current password and
// updating the password with a new one.
func (h *userHandler) editPostHandler(w http.ResponseWriter, r *http.Request) {
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
}

// signoutHandler is a handler function that handles the user's signout request.
// It clears the session data related to authenticated user.
func (h *userHandler) signoutHandler(w http.ResponseWriter, r *http.Request) {
	h.CookieSession.DestroySession(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
