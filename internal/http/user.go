package http

import (
	"log"
	"net/http"

	"github.com/mnlg/seonaut/internal/helper"
	"github.com/mnlg/seonaut/internal/user"
)

func (app *App) serveSignup(w http.ResponseWriter, r *http.Request) {
	var e bool
	var email string

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("serveSignup ParseForm: %v\n", err)
			http.Redirect(w, r, "/signup", http.StatusSeeOther)

			return
		}

		email = r.FormValue("email")
		password := r.FormValue("password")

		err = app.userService.SignUp(email, password)
		if err == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		log.Printf("serveSignup SignUp: %v\n", err)
		e = true
	}

	app.renderer.RenderTemplate(w, "signup", &helper.PageView{
		PageTitle: "SIGNUP_VIEW",
		Data: struct {
			Email string
			Error bool
		}{Error: e, Email: email},
	})
}

func (app *App) serveSignin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("serveSignin ParseForm: %v\n", err)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)

			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		u, err := app.userService.SignIn(email, password)
		if err != nil {
			log.Printf("serveSignin SignIn: %v\n", err)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		session, _ := app.cookie.Get(r, "SESSION_ID")
		session.Values["authenticated"] = true
		session.Values["uid"] = u.Id
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	v := &helper.PageView{
		PageTitle: "SIGNIN_VIEW",
	}

	app.renderer.RenderTemplate(w, "signin", v)
}

func (app *App) serveSignout(user *user.User, w http.ResponseWriter, r *http.Request) {
	session, _ := app.cookie.Get(r, "SESSION_ID")
	session.Values["authenticated"] = false
	session.Values["uid"] = nil
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *App) requireAuth(f func(user *user.User, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := app.cookie.Get(r, "SESSION_ID")
		var authenticated interface{} = session.Values["authenticated"]
		if authenticated != nil {
			isAuthenticated := session.Values["authenticated"].(bool)
			if isAuthenticated {
				session, _ := app.cookie.Get(r, "SESSION_ID")
				uid := session.Values["uid"].(int)
				user := app.userService.FindById(uid)
				f(user, w, r)

				return
			}
		}

		http.Redirect(w, r, "/signin", http.StatusSeeOther)
	}
}
