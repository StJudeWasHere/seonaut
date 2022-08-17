package http

import (
	"log"
	"net/http"
)

func (app *App) serveSignup(w http.ResponseWriter, r *http.Request) {
	var email string

	pageData := &struct {
		Email string
		Error bool
	}{}

	pageView := &PageView{
		PageTitle: "SIGNUP_VIEW",
		Data:      pageData}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("serveSignup ParseForm: %v\n", err)
			http.Redirect(w, r, "/signup", http.StatusSeeOther)

			return
		}

		email = r.FormValue("email")
		password := r.FormValue("password")

		pageData.Email = email

		err = app.userService.SignUp(email, password)
		if err != nil {
			log.Printf("serveSignup SignUp: %v\n", err)
			pageData.Error = true
			app.renderer.RenderTemplate(w, "signup", pageView)
			return
		}

		u, err := app.userService.SignIn(email, password)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
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

func (app *App) serveSignin(w http.ResponseWriter, r *http.Request) {
	var e bool
	var email string

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("serveSignin ParseForm: %v\n", err)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)

			return
		}

		email = r.FormValue("email")
		password := r.FormValue("password")

		u, err := app.userService.SignIn(email, password)
		if err == nil {
			session, _ := app.cookie.Get(r, "SESSION_ID")
			session.Values["authenticated"] = true
			session.Values["uid"] = u.Id
			session.Save(r, w)

			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		log.Printf("serveSignin SignIn: %v\n", err)
		e = true
	}

	v := &PageView{
		PageTitle: "SIGNIN_VIEW",
		Data: struct {
			Email string
			Error bool
		}{Email: email, Error: e},
	}

	app.renderer.RenderTemplate(w, "signin", v)
}

func (app *App) serveAccount(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Error        bool
		ErrorMessage string
	}{}

	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
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

			v := &PageView{
				PageTitle: "ACCOUNT_VIEW",
				Data:      data,
			}

			app.renderer.RenderTemplate(w, "account", v)
			return
		}

		err = app.userService.UpdatePassword(user.Email, newPassword)
		if err != nil {
			data.Error = true
			data.ErrorMessage = "New password is not valid."

			v := &PageView{
				PageTitle: "ACCOUNT_VIEW",
				Data:      data,
			}

			app.renderer.RenderTemplate(w, "account", v)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	v := &PageView{
		PageTitle: "ACCOUNT_VIEW",
		Data:      data,
	}

	app.renderer.RenderTemplate(w, "account", v)
}

func (app *App) serveSignout(w http.ResponseWriter, r *http.Request) {
	session, _ := app.cookie.Get(r, "SESSION_ID")
	session.Values["authenticated"] = false
	session.Values["uid"] = nil
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

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
