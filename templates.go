package main

import (
	"html/template"
	"log"
	"net/http"
)

func renderTemplate(w http.ResponseWriter, t string, d interface{}) {
	var templates = template.Must(
		template.ParseFiles(
			"templates/footer.html",
			"templates/head.html",
			"templates/home.html",
			"templates/issues_view.html",
			"templates/issues.html",
			"templates/list.html",
			"templates/pagereport.html",
			"templates/project_add.html",
			"templates/resources.html",
			"templates/signin.html",
			"templates/signup.html",
		))

	err := templates.ExecuteTemplate(w, t+".html", d)
	if err != nil {
		log.Println(err)
	}
}
