package main

import (
	"html/template"
	"image/png"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/qr"
)

var t *template.Template

type ViewState struct {
	Username  string
	Config    Config
	Users     UserList
	Plants    PlantList
	Varieties VarietyList
	PageTitle string
	PageName  string
	Page      Page
	ID        int
	Objects   []Object
	Session   Session
}

type GetParam struct {
	Key   string
	Value string
}

func startWebserver() {
	logMessage("Core", "Started Webserver on port "+config.Webserver.Port)

	t = loadPages()

	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/doLoginAction", doLogin)
	http.HandleFunc("/home/", renderPage)
	http.HandleFunc("/home/varieties/createVariety", doCreateVariety)
	http.HandleFunc("/home/users/createUser", doCreateUser)
	http.HandleFunc("/barcode/", genBarcode)
	http.HandleFunc("/label/", genLabel)
	http.HandleFunc("/admin/passwordChange/", passChange)
	http.HandleFunc("/admin/deleteAccount/", doDeleteUser)
	http.HandleFunc("/admin/passwordChange/doPasswordChangeAction", doPasswordChange)
	http.HandleFunc("/api/trefle/query/", trefleQuery)
	http.HandleFunc("/api/trefle/plant/id/", treflePlantByID)

	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./templates/js"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./templates/img"))))
	http.HandleFunc("/manifest.webmanifest", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./manifest.webmanifest")
	})
	log.Fatal(http.ListenAndServeTLS(":"+config.Webserver.Port, "server.crt", "server.key", nil))
}

func landingPage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "./login", 301)
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	_, valid := checkLogin(w, r)
	if valid {
		http.Redirect(w, r, "https://"+config.Webserver.Hostname+":"+config.Webserver.Port+"/home/", 302)
		return
	}
	w.WriteHeader(200)
	t.ExecuteTemplate(w, "login.html", nil)
}

func passChange(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/admin/passwordChange/"):]
	var state ViewState
	state.Users = getAllUsers()
	state.ID, _ = strconv.Atoi(strings.Split(id, "?")[0])
	state.Config = config

	session, _ := restrictPage(w, r, true, true)

	if len(r.URL.Path[len("/admin/passwordChange/"):]) > 3 {
		var redir GetParam
		redir.Key = strings.Split(strings.Split(id, "?")[1], "=")[0]
		redir.Value = strings.Split(strings.Split(id, "?")[1], "=")[1]
		if redir.Key == "r" {
			session.Redirect = redir.Value
		}
	}
	w.WriteHeader(200)
	t.ExecuteTemplate(w, "passwordChange.html", state)
}

func genLabel(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/label/p/"):]
	var list PlantList
	strList := strings.Split(id, "/")
	for i := 0; i < len(strList); i++ {
		var p Plant
		p.PlantID, _ = strconv.Atoi(strList[i])
		p = getPlantData(p)
		list.Plants = append(list.Plants, p)
	}

	var state ViewState
	state.Plants = list
	state.Config = config

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html") // <-- set the content-type header
	t.ExecuteTemplate(w, config.Hardware.LabelMaker.Barcode+"-"+config.Hardware.LabelMaker.LabelSize+".html", state)
}

func genBarcode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png") // <-- set the content-type header
	content := r.URL.Path[len("/barcode/"):]
	params := strings.Split(content, "/")
	if params[0] == "code128" {
		barcode, _ := code128.Encode(params[1])
		png.Encode(w, barcode)
	} else if params[0] == "qr" {
		barcode, _ := qr.Encode(params[1], qr.M, qr.Auto)
		png.Encode(w, barcode)
	}
}

func doLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		panic(err)
	}
	user := r.Form.Get("username")
	pass := r.Form.Get("password")
	redir := r.Form.Get("redirect")

	var userObj, loginState = loginUser(user, pass)

	if userObj.Hash == "change" {
		registerSessionWithClient(w, userObj.UserID)
		http.Redirect(w, r, "http://"+config.Webserver.Hostname+":"+config.Webserver.Port+"/admin/passwordChange/"+strconv.Itoa(userObj.UserID)+"?r=/home/firstRun", 302)
	}

	if loginState {
		logMessage("Core", "User "+user+" logged in successfully")
		registerSessionWithClient(w, userObj.UserID)
		if redir != "" {
			decodedValue, _ := url.QueryUnescape(redir)
			http.Redirect(w, r, "."+decodedValue, 302)
		} else {
			http.Redirect(w, r, "./home", 302)
		}
		return
	} else {
		logMessage("Core", "Failed login attempt of "+user)
		http.Redirect(w, r, "./login?e=1", 302)
		return
	}

	//body, _ := ioutil.ReadFile("../templates/login.html")
	//fmt.Fprint(w, string(body))
}

func doPasswordChange(w http.ResponseWriter, r *http.Request) {
	session, _ := restrictPage(w, r, true, true)
	err := r.ParseForm()
	if err != nil {
		panic(err)
	}
	userID := r.Form.Get("userID")
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	uID, _ := strconv.Atoi(userID)

	changeUserPass(uID, password)
	logMessage("Core", "Changed "+username+"'s password")
	if isRedirectPending(session) {
		doRedirect(w, r, session)
	} else {
		http.Redirect(w, r, "../../../home/user/"+userID, 302)
	}
}

func doCreateVariety(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		panic(err)
	}
	var variety Variety
	variety.VarietyName = r.Form.Get("name")
	variety.VarietyFamily = r.Form.Get("family")
	variety.VarietyCatagory = r.Form.Get("catagory")
	variety.SunlightCatagory = r.Form.Get("sun")
	variety.WaterCatagory = r.Form.Get("water")
	variety.AvgSprout, _ = strconv.Atoi(r.Form.Get("sprout"))
	variety.AvgHarvest, _ = strconv.Atoi(r.Form.Get("harvest"))
	createVariety(variety)
	logMessage("Core", "Added Variety: "+r.Form.Get("name"))
	http.Redirect(w, r, "../varieties", 302)
}

func doCreateUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		panic(err)
	}
	var user User
	user.Username = r.Form.Get("username")
	user = createUser(user)
	logMessage("Core", "Added User: "+r.Form.Get("name"))
	http.Redirect(w, r, "http://"+config.Webserver.Hostname+":"+config.Webserver.Port+"/admin/passwordChange/"+strconv.Itoa(user.UserID), 302)
}

func doDeleteUser(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/admin/deleteAccount/"):]
	userID, _ := strconv.Atoi(title)
	deleteUser(userID)
	http.Redirect(w, r, "http://"+config.Webserver.Hostname+":"+config.Webserver.Port+"/home/users/", 302)
}

func doRedirect(w http.ResponseWriter, r *http.Request, s Session) {
	redir := s.Redirect
	s.Redirect = ""
	http.Redirect(w, r, "http://"+config.Webserver.Hostname+":"+config.Webserver.Port+redir, 302)
}

func isRedirectPending(s Session) bool {
	return s.Redirect != ""
}
