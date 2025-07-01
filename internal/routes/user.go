package routes

import (
	"log"
	"net/http"

	"github.com/stjudewashere/seonaut/internal/services"
)

type userHandler struct {
	*services.Container
}

// signupGetHandler handles the GET signup request and displays the sign up form.
// It allows users to sign up by providing their email and password.
func (h *userHandler) signupGetHandler(w http.ResponseWriter, r *http.Request) {
	pageView := &PageView{
		PageTitle: "SIGNUP_VIEW_PAGE_TITLE",
		Data: &struct {
			Email        string
			Error        bool
			ErrorMessage string
		}{},
	}

	h.Renderer.RenderTemplate(w, "signup", pageView)
}

// signupPostHandler handles the POST request of the signup form.
// Upon successful signup, the user is automatically signed in and redirected to the home page.
// In case of error the signup template is rendered with the pre-populated form using the
// data's email field.
func (h *userHandler) signupPostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	u, err := h.UserService.SignUp(email, password)
	if err != nil {
		errorMsg := "The email address or password is not valid."
		switch err {
		case services.ErrInvalidPassword:
			errorMsg = "Password is not valid."
		case services.ErrInvalidEmail:
			errorMsg = "Email address is not valid."
		default:
			log.Printf("sign up error: %v", err)
		}
		pageView := &PageView{
			PageTitle: "SIGNUP_VIEW_PAGE_TITLE",
			Data: &struct {
				Email        string
				Error        bool
				ErrorMessage string
			}{
				Email:        email,
				Error:        true,
				ErrorMessage: errorMsg,
			},
		}

		h.Renderer.RenderTemplate(w, "signup", pageView)
		return
	}

	h.CookieSession.SetSession(u, w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// deleteGetHandler handles the HTTP GET request for the delete user account functionality.
// it renders the delete page with the appropriate data.
// A user account can not be deleted if it is crawling a project.
func (h *userHandler) deleteGetHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	pageView := &PageView{
		PageTitle: "DELETE_ACCOUNT_VIEW_PAGE_TITLE",
		User:      *user,
		Data: &struct {
			Crawling bool
		}{
			Crawling: h.ProjectViewService.UserIsCrawling(user.Id),
		},
	}

	h.Renderer.RenderTemplate(w, "delete_account", pageView)
}

// deletePostHandler handles the POST request to delete an account.
// After deleting the user and all its associated data it destroys the session
// and redirects to the sign-in page. Users with projects still crawling can not
// be deleted and are redirected to the delete page.
func (h *userHandler) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	crawling := h.ProjectViewService.UserIsCrawling(user.Id)
	if crawling {
		http.Redirect(w, r, "/account/delete", http.StatusSeeOther)
		return
	}

	h.UserService.DeleteUser(user)
	h.CookieSession.DestroySession(w, r)
	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}

// signinGetHandler handles the HTTP GET request for the sign-in functionality.
// It renders the sign-in page with the appropriate data.
// The signin form is pre-populated using the data's email field.
func (h *userHandler) signinGetHandler(w http.ResponseWriter, r *http.Request) {
	pageView := &PageView{
		PageTitle: "SIGNIN_VIEW_PAGE_TITLE",
		Data: &struct {
			Email        string
			Error        bool
			ErrorMessage string
		}{},
	}

	h.Renderer.RenderTemplate(w, "signin", pageView)
}

// signinGetHandler handles the HTTP POST request for the sign-in functionality.
// It validates the user's credentials, and if the sign in is successful, it creates a session
// and redirects the user to the projects homepage.
// The signin form is pre-populated using the data's email field.
// In case of error, the signin page is rendered and the data's Error field is set to true.
func (h *userHandler) signinPostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	u, err := h.UserService.SignIn(email, password)
	if err != nil {
		pageView := &PageView{
			PageTitle: "SIGNIN_VIEW_PAGE_TITLE",
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

// editGetHandler handles the HTTP GET request for the account edit page.
// It renders the account management form with the appropriate data.
func (h *userHandler) editGetHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	pageView := &PageView{
		PageTitle: "ACCOUNT_VIEW_PAGE_TITLE",
		User:      *user,
		Data: &struct {
			Error        bool
			ErrorMessage string
		}{},
	}

	h.Renderer.RenderTemplate(w, "account", pageView)
}

// editPostHandler handles the HTTP POST request for the account edit page.
// It allows users to change their credentials by verifying the current password and
// updating it with a new one.
// In case of error the data's Error and ErrorMessage are populated for the template.
func (h *userHandler) editPostHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	currentPassword := r.FormValue("password")
	newPassword := r.FormValue("new_password")

	err = h.UserService.UpdatePassword(user, currentPassword, newPassword)
	if err != nil {
		errorMsg := "An error occurred. Please try again."
		switch err {
		case services.ErrInvalidPassword:
			errorMsg = "New password is not valid."
		case services.ErrIncorrectPassword:
			errorMsg = "Current password is incorrect."
		default:
			log.Printf("update password user id %d error: %v", user.Id, err)
		}

		pageView := &PageView{
			PageTitle: "ACCOUNT_VIEW_PAGE_TITLE",
			User:      *user,
			Data: &struct {
				Error        bool
				ErrorMessage string
			}{
				Error:        true,
				ErrorMessage: errorMsg,
			},
		}

		h.Renderer.RenderTemplate(w, "account", pageView)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// signoutHandler handles the user's signout request.
// It clears the session data related to authenticated user and redirects to
// the sign-in page.
func (h *userHandler) signoutHandler(w http.ResponseWriter, r *http.Request) {
	h.CookieSession.DestroySession(w, r)
	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}
