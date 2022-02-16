package main

import (
	"html/template"
	"log"
	"net/http"
)

type PageView struct {
	PageTitle string
	User      User
	Data      interface{}
}

func renderTemplate(w http.ResponseWriter, t string, v *PageView) {
	var templates = template.Must(
		template.ParseFiles(
			"templates/footer.html",
			"templates/head.html",
			"templates/home.html",
			"templates/issues_view.html",
			"templates/issues.html",
			"templates/charts.html",
			"templates/project_add.html",
			"templates/resources.html",
			"templates/signin.html",
			"templates/signup.html",
			"templates/upgrade.html",
			"templates/manage.html",
			"templates/canceled.html",
		))

	err := templates.ExecuteTemplate(w, t+".html", v)
	if err != nil {
		log.Println(err)
	}
}
