package main

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Page struct {
	NameTitle           string //Frontend Name
	Name                string //Template Name
	URLPattern          string //URL Pattern
	Internal            bool   //Require User to be Logged In
	RequiresPermissions string //Permissions Required to View Page
}

func LoadPages() (t *template.Template) {
	t = template.Must(template.New("template").Funcs(template.FuncMap{
		"safe": func(s string) template.HTML { return template.HTML(s) },
	}), nil)

	LogMessage("Loading avalible templates")
	filepath.Walk("./templates", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".html") {
			LogMessage("Found template: " + path)
			var _, er = t.ParseFiles(path)
			if er != nil {
				log.Fatal(er)
			}
			if len(path) > len("templates/pages/.html") && path[len("templates/"):len("templates/pages")] == "pages" {

				//pageName := path[len("templates/pages/") : len(path)-+len(".html")]

				//var p Page

			}
		}

		return nil
	})
	return t
}
