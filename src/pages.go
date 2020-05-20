package main

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
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
	RequireLoggedIn   bool
	RequirePermission string
	PullPlants        bool
	PullUsers         bool
	PullVarieties     bool
	PageTitle         string
	PageName          string
	URLPattern        string
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

func readPageFlags(key string, value string, f PageFlags) (flags PageFlags, v bool) {

	e := reflect.ValueOf(&f).Elem()

	v = false

	for i := 0; i < e.NumField(); i++ {
		if key == e.Type().Field(i).Name {
			switch e.Type().Field(i).Type {
			case reflect.TypeOf((*string)(nil)).Elem():
				e.Field(i).SetString(value)
				v = true
			case reflect.TypeOf((*int)(nil)).Elem():
				tmp, _ := strconv.Atoi(value)
				e.Field(i).SetInt(int64(tmp))
				v = true
			case reflect.TypeOf((*bool)(nil)).Elem():
				if value == "true" {
					e.Field(i).SetBool(true)
					v = true
				} else if value == "true" {
					e.Field(i).SetBool(false)
					v = true
				}

			}
		}
	}
	flags = f
	return
}

func renderPage(w http.ResponseWriter, r *http.Request) {

	var state ViewState
	state.Config = config
	requestPath := r.URL.Path[len("/home/"):]
	logMessage("Core", "Serving "+r.URL.Path)
	if r.URL.Path == "/home" || r.URL.Path == "/home/" {
		redirectPage(w, r, "/home/dashboard")
		return
	}
	for i := 0; i < len(SystemPages); i++ {
		requestParams := strings.Split(requestPath, "/")
		if requestParams[0] == SystemPages[i].URLPattern {
			if len(requestParams) > 1 {
				state.ID, _ = strconv.Atoi(requestParams[1])
			} else {
				state.ID = 0
			}
			state.Session, _ = restrictPage(w, r, SystemPages[i].Flags.RequireLoggedIn, SystemPages[i])
			state.PageTitle = SystemPages[i].PageTitle
			state.PageName = SystemPages[i].PageName
			state.Objects = SystemObjects

			if SystemPages[i].Flags.PullPlants {
				if state.Session.IsAdmin {
					logMessage("DB", "Pulled global plant list")
					state.Plants = getAllPlants(-1)
				} else {
					logMessage("DB", "Pulled plant list for user "+state.Session.Username)
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

	logMessage("Templates", "Loading avalible templates")
	filepath.Walk("./templates", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".html") {
			logMessage("Templates", "Added template: "+path)
			var _, er = t.ParseFiles(path)
			if er != nil {
				log.Fatal(er)
			}
			if len(path) > len("templates/pages/.html") && path[len("templates/"):len("templates/pages")] == "pages" {

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

		rawFlagBlock := strings.Split(templateCode, "{{safe \"<!--ENDPAGEFLAGS!-->\"}}")[0]
		flagBlock := re.ReplaceAllString(rawFlagBlock, "")
		flagPairs := strings.Split(flagBlock, ",")

		var o Object
		for i := 0; i < len(flagPairs); i++ {
			flagPair := strings.Split(flagPairs[i], ": ")
			var v bool
			f, v = readPageFlags(flagPair[0], flagPair[1], f)
			if !v {
				switch flagPair[0] {
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

func redirectPage(w http.ResponseWriter, r *http.Request, path string) {
	http.Redirect(w, r, "https://"+config.Webserver.Hostname+":"+config.Webserver.Port+path, 302)
}
