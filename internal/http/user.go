package http

import (
	"log"
	"net/http"
)

// handleSignup handles the signup functionality for the application.
// It allows users to sign up by providing their email and password.
// Upon successful signup, the user is automatically signed in and redirected to the home page.
//
// The function handles both GET and POST HTTP methods.
// GET: Renders the signup form.
// POST: Processes the signup form data, performs signup, signs the user in, and redirects to the home page.
func (app *App) handleSignup(w http.ResponseWriter, r *http.Request) {
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

		u, err := app.userService.SignUp(data.Email, password)
		if err != nil {
			log.Printf("serveSignup SignUp: %v", err)
			data.Error = true
			app.renderer.RenderTemplate(w, "signup", pageView)

			return
		}

		session, _ := app.cookie.Get(r, "SESSION_ID")
		session.Values["authenticated"] = true
		session.Values["uid"] = u.Id
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	app.renderer.RenderTemplate(w, "signup", pageView)
}

// handleDeleteUser handles the HTTP GET and POST requests for the delete user account functionality.
//
// The function handles both GET and POST HTTP methods.
// GET: it renders the delete page with the appropriate data.
// POST: it sign's out the user and deletes the account including all its associated data.
func (app *App) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	user, ok := app.userService.GetUserFromContext(r.Context())
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

	pv := app.projectViewService.GetProjectViews((user.Id))
	for _, v := range pv {
		if !v.Crawl.IssuesEnd.Valid || v.Project.Deleting {
			data.Crawling = true
			app.renderer.RenderTemplate(w, "delete_account", pageView)
			return
		}
	}

	if r.Method == http.MethodPost {
		session, _ := app.cookie.Get(r, "SESSION_ID")
		session.Values["authenticated"] = false
		session.Values["uid"] = nil
		session.Save(r, w)

		app.userService.DeleteUser(user)

		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	app.renderer.RenderTemplate(w, "delete_account", pageView)
}

// handleSignin handles the HTTP GET and POST requests for the sign-in functionality.
//
// The function handles both GET and POST HTTP methods.
// GET: it renders the sign-in page with the appropriate data.
// POST: it validates the user's credentials and creates a session if the sign-in is successful.
func (app *App) handleSignin(w http.ResponseWriter, r *http.Request) {
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

		u, err := app.userService.SignIn(data.Email, password)
		if err == nil {
			session, _ := app.cookie.Get(r, "SESSION_ID")
			session.Values["authenticated"] = true
			session.Values["uid"] = u.Id
			session.Save(r, w)

			http.Redirect(w, r, "/", http.StatusSeeOther)

			return
		}

		log.Printf("serveSignin SignIn: %v\n", err)
		data.Error = true
	}

	app.renderer.RenderTemplate(w, "signin", pageView)
}

// handleAccount handles the HTTP POST and GET requests for the account management functionality.
//
// The function handles both GET and POST HTTP methods.
// POST: it allows users to change their credentials by verifying the current password and
// updating the password with a new one.
// GET: it renders the account management form with the appropriate data.
func (app *App) handleAccount(w http.ResponseWriter, r *http.Request) {
	user, ok := app.userService.GetUserFromContext(r.Context())
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

		_, err = app.userService.SignIn(user.Email, password)
		if err != nil {
			data.Error = true
			data.ErrorMessage = "Current password is not correct."

			app.renderer.RenderTemplate(w, "account", pageView)

			return
		}

		err = app.userService.UpdatePassword(user.Email, newPassword)
		if err != nil {
			data.Error = true
			data.ErrorMessage = "New password is not valid."

			app.renderer.RenderTemplate(w, "account", pageView)

			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	app.renderer.RenderTemplate(w, "account", pageView)
}

// handleSignout is a handler function that handles the user's signout request.
// It clears the session data related to authenticated user.
func (app *App) handleSignout(w http.ResponseWriter, r *http.Request) {
	session, _ := app.cookie.Get(r, "SESSION_ID")
	session.Values["authenticated"] = false
	session.Values["uid"] = nil
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// requireAuth is a middleware function that wraps the provided handler function and enforces authentication.
// It checks if the user is authenticated based on the session data.
func (app *App) requireAuth(f func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := app.cookie.Get(r, "SESSION_ID")
		var authenticated interface{} = session.Values["authenticated"]
		if authenticated != nil {
			isAuthenticated := session.Values["authenticated"].(bool)
			if isAuthenticated {
				session, _ := app.cookie.Get(r, "SESSION_ID")
				session.Options.MaxAge = 60 * 60 * 48
				session.Save(r, w)
				uid := session.Values["uid"].(int)
				user := app.userService.FindById(uid)
				ctx := app.userService.SetUserToContext(user, r.Context())
				req := r.WithContext(ctx)
				f(w, req)

				return
			}
		}

		http.Redirect(w, r, "/signin", http.StatusSeeOther)
	}
}
