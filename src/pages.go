package main

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type Page struct {
	PageTitle  string
	PageName   string
	URLPattern string
	Flags      PageFlags
	Code       template.HTML
}

type PageFlags struct {
	RequireLoggedIn bool
	RequireAdmin    bool
	PullPlants      bool
	PullUsers       bool
	PullVarieties   bool
	PageTitle       string
	PageName        string
	URLPattern      string
}

type Object struct {
	Name        string
	DisplayName string
	IconName    string
}

type Hook struct {
	PageName      string
	OnLogin       func(w http.ResponseWriter, r *http.Request, s Session)
	OnPageRequest func(w http.ResponseWriter, r *http.Request, s Session)
}

var SystemPages []Page
var SystemHooks []Hook
var SystemObjects []Object

func registerObject(o Object) {
	if !contains(SystemObjects, o) {
		SystemObjects = append(SystemObjects, o)
	}
}

func registerHook(h Hook) {
	SystemHooks = append(SystemHooks, h)
}

func callPageRequestHooks(w http.ResponseWriter, r *http.Request, s Session) {
	for i := 0; i < len(SystemHooks); i++ {
		if SystemHooks[i].OnPageRequest != nil {
			SystemHooks[i].OnPageRequest(w, r, s)
		}
	}
}

func registerPage(p Page) {
	SystemPages = append(SystemPages, p)
}

func renderPage(w http.ResponseWriter, r *http.Request) {

	var state ViewState
	state.Config = config

	requestPath := r.URL.Path[len("/home/"):]
	for i := 0; i < len(SystemPages); i++ {
		requestParams := strings.Split(requestPath, "/")
		if requestParams[0] == SystemPages[i].URLPattern {
			if len(requestParams) > 1 {
				state.ID, _ = strconv.Atoi(requestParams[1])
			} else {
				state.ID = 0
			}
			state.Session, _ = restrictPage(w, r, SystemPages[i].Flags.RequireLoggedIn, SystemPages[i].Flags.RequireAdmin)
			state.PageTitle = SystemPages[i].PageTitle
			state.PageName = SystemPages[i].PageName
			state.Objects = SystemObjects
			if SystemPages[i].Flags.PullPlants {
				if state.Session.IsAdmin {
					logMessage("core", "Pulled Global Plant List From DB")
					state.Plants = getAllPlants(-1)
				} else {
					logMessage("core", "Pulled Plant List From DB")
					state.Plants = getAllPlants(state.Session.UserID)
				}
			}
			if SystemPages[i].Flags.PullVarieties {
				state.Varieties = getAllVarieties()
			}
			if SystemPages[i].Flags.PullUsers {
				state.Users = getAllUsers()
			}

			callPageRequestHooks(w, r, state.Session)

			w.WriteHeader(200)

			buf := new(bytes.Buffer)
			t.ExecuteTemplate(buf, state.PageName+".html", state)

			state.Page.Code = template.HTML(string(template.HTML(buf.String())))

			if strings.Contains(string(state.Page.Code), "<!--ENDPAGEFLAGS!-->") {
				state.Page.Code = template.HTML(strings.Split(string(state.Page.Code), "<!--ENDPAGEFLAGS!-->")[1])
			}
			t.ExecuteTemplate(w, "root.html", state)
		}
	}
}

func loadPages() (t *template.Template) {
	t = template.Must(template.New("template").Funcs(template.FuncMap{
		"safe": func(s string) template.HTML { return template.HTML(s) },
	}), nil)

	filepath.Walk("./templates", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".html") {
			logMessage("Core", "Added template: "+path)
			var _, er = t.ParseFiles(path)
			if er != nil {
				log.Fatal(er)
			}
			if len(path) > len("templates/pages/.html") && path[len("templates/"):len("templates/pages")] == "pages" {
				logMessage("core", "detected page: "+path[len("templates/pages/"):len(path)-+len(".html")])

				pageName := path[len("templates/pages/") : len(path)-+len(".html")]

				var p Page
				p.Flags = getPageFlags(path)

				if p.Flags.PageName == "" {
					p.PageName = pageName
				} else {
					p.PageName = p.Flags.PageName
				}

				if p.Flags.PageTitle == "" {
					p.PageTitle = camelCaseConv(pageName)
				} else {
					p.PageTitle = p.Flags.PageTitle
				}

				if p.Flags.URLPattern == "" {
					p.URLPattern = pageName
				} else {
					p.URLPattern = p.Flags.URLPattern
				}
				registerPage(p)
			}
		}

		return nil
	})
	return t
}

func getPageFlags(path string) (f PageFlags) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	// Convert []byte to string and print to screen
	templateCode := string(content)

	//Did Template Declare Flags?
	if strings.Contains(templateCode, "<!--ENDPAGEFLAGS!-->") {
		re := regexp.MustCompile(`\r?\n`)

		rawFlagBlock := strings.Split(templateCode, "<!--ENDPAGEFLAGS!-->")[0]
		flagBlock := re.ReplaceAllString(rawFlagBlock, "")
		flagPairs := strings.Split(flagBlock, ",")

		var o Object
		for i := 0; i < len(flagPairs); i++ {
			flagPair := strings.Split(flagPairs[i], ": ")
			switch flagPair[0] {
			case "RequireLoggedIn":
				f.RequireLoggedIn, _ = strconv.ParseBool(flagPair[1])
			case "RequireAdmin":
				f.RequireAdmin, _ = strconv.ParseBool(flagPair[1])
			case "PullPlants":
				f.PullPlants, _ = strconv.ParseBool(flagPair[1])
			case "PullUsers":
				f.PullUsers, _ = strconv.ParseBool(flagPair[1])
			case "PullVarieties":
				f.PullVarieties, _ = strconv.ParseBool(flagPair[1])
			case "PageTitle":
				f.PageTitle = flagPair[1]
			case "PageName":
				f.PageName = flagPair[1]
			case "URLPattern":
				f.URLPattern = flagPair[1]
			case "ObjectName":
				o.Name = flagPair[1]
				o.DisplayName = strings.Title(flagPair[1])
			case "DisplayName":
				o.DisplayName = strings.Title(flagPair[1])
			case "ObjectIcon":
				o.IconName = flagPair[1]
			default:
				logMessage("core", "Unknown page flag '"+flagPair[0]+"' in template '"+path+"'")
			}
		}
		if o.Name != "" {
			registerObject(o)
		}
	}
	return
}

func camelCaseConv(s string) string {
	buf := &bytes.Buffer{}
	for i, rune := range s {
		if unicode.IsUpper(rune) && i > 0 {
			buf.WriteRune(' ')
		}
		buf.WriteRune(rune)
	}
	return strings.Title(buf.String())
}

func contains(s []Object, e Object) bool {
	for _, a := range s {
		if a.Name == e.Name {
			return true
		}
	}
	return false
}
