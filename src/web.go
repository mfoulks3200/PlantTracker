package main

import (
	"html/template"
	"image/png"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/boombuler/barcode/code128"
)

var t *template.Template

type ViewState struct {
	Username  string
	Config    Config
	Users     UserList
	PageTitle string
	Page      string
	ID        int
}

type LabelState struct {
	PlantID            int
	PlantName          string
	PlantVariety       string
	PlantVarietyFamily string
	PlantCatagory      string
	SunlightCatagory   string
	WaterCatagory      string
	AvgSprout          string
	AvgHarvest         string
	PlantDate          string
}

func startWebserver() {
	logMessage("Core", "Started Webserver on port "+config.Webserver.Port)

	t = template.New("template")

	filepath.Walk("./templates", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".html") {
			logMessage("Core", "Added template: "+path)
			var _, er = t.ParseFiles(path)
			if er != nil {
				log.Fatal(er)
			}
		}

		return nil
	})

	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/doLoginAction", doLogin)
	http.HandleFunc("/home/", homePage)
	http.HandleFunc("/barcode/", genBarcode)
	http.HandleFunc("/label/", genLabel)
	log.Fatal(http.ListenAndServe(":"+config.Webserver.Port, nil))
}

func landingPage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "./login", 301)
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	t.ExecuteTemplate(w, "login.html", nil)
}

func genLabel(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/label/p/"):]
	intID, _ := strconv.Atoi(id)
	var plant = getPlant(intID)
	if plant.PlantID != 0 {
		intVID, _ := strconv.Atoi(plant.PlantVariety)
		var variety = getVariety(intVID)
		var state LabelState
		state.PlantID = plant.PlantID
		state.PlantName = plant.PlantName
		state.PlantVariety = variety.VarietyName
		state.PlantVarietyFamily = variety.VarietyFamily
		state.PlantCatagory = variety.VarietyCatagory
		state.SunlightCatagory = variety.SunlightCatagory
		state.WaterCatagory = variety.WaterCatagory
		const longForm = "1/02/06Z03:04:05"
		pDate, _ := time.Parse(longForm, plant.PlantDate)

		sDate := pDate.Add(time.Duration(variety.AvgSprout) * time.Hour * 24)
		state.AvgSprout = strings.Split(sDate.Format("01/02/06"), "Z")[0]

		hDate := pDate.Add(time.Duration(variety.AvgHarvest) * time.Hour * 24)
		state.AvgHarvest = strings.Split(hDate.Format("01/02/06"), "Z")[0]

		s := strings.Split(plant.PlantDate, "Z")
		state.PlantDate = s[0]
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "text/html") // <-- set the content-type header
		t.ExecuteTemplate(w, config.Hardware.LabelMaker.LabelSize+".html", state)
	} else {
		w.WriteHeader(404)
	}
}

func genBarcode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png") // <-- set the content-type header
	content := r.URL.Path[len("/barcode/"):]
	barcode, _ := code128.Encode(content)
	png.Encode(w, barcode)
}

func homePage(w http.ResponseWriter, r *http.Request) {

	title := r.URL.Path[len("/home/"):]
	var state ViewState
	if len(title) != 0 {
		if title[0:5] == "users" {
			state.Users = getAllUsers()
			state.PageTitle = "Users"
			state.Page = "users"
			state.Username = "mfoulks200"
			state.Config = config
			state.ID = 0
			w.WriteHeader(200)
			t.ExecuteTemplate(w, "root.html", state)
		} else if title[0:4] == "user" {
			state.Users = getAllUsers()
			state.PageTitle = "Users"
			state.Page = "user"
			state.Username = "mfoulks200"
			state.Config = config
			state.ID, _ = strconv.Atoi(title[6:])
			w.WriteHeader(200)
			t.ExecuteTemplate(w, "root.html", state)
		} else if title[0:5] == "plants" {
			state.PageTitle = "Users"
			state.Page = "users"
			state.Username = "mfoulks200"
			state.Config = config
			state.ID = 0
			w.WriteHeader(200)
			t.ExecuteTemplate(w, "root.html", state)
		}
	} else {
		state.PageTitle = "Dashboard"
		state.Page = "dashboard"
		state.Username = "mfoulks200"
		state.Config = config
		state.ID = 0
		w.WriteHeader(200)
		t.ExecuteTemplate(w, "root.html", state)
	}

}

func doLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		panic(err)
	}
	user := r.Form.Get("username")
	pass := r.Form.Get("password")

	var loginState = loginUser(user, pass)

	if loginState {
		logMessage("Core", "User "+user+" logged in successfully")
		http.Redirect(w, r, "./home", 302)
		return
	} else {
		logMessage("Core", "Failed login attempt of "+user)
		http.Redirect(w, r, "./login?e=1", 302)
		return
	}

	//body, _ := ioutil.ReadFile("../templates/login.html")
	//fmt.Fprint(w, string(body))
}
