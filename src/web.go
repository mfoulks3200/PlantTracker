package main

import (
	"html/template"
	"image/png"
	"log"
	"net/http"
	"net/url"
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
	Plants    PlantList
	Varieties VarietyList
	PageTitle string
	Page      string
	ID        int
	Session   Session
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
	http.HandleFunc("/home/varieties/createVariety", doCreateVariety)
	http.HandleFunc("/home/users/createUser", doCreateUser)
	http.HandleFunc("/barcode/", genBarcode)
	http.HandleFunc("/label/", genLabel)
	http.HandleFunc("/admin/passwordChange/", passChange)
	http.HandleFunc("/admin/deleteAccount/", doDeleteUser)
	http.HandleFunc("/admin/passwordChange/doPasswordChangeAction", doPasswordChange)
	http.HandleFunc("/api/trefle/query/", trefleQuery)
	http.HandleFunc("/api/trefle/plant/id/", treflePlantByID)
	log.Fatal(http.ListenAndServe(":"+config.Webserver.Port, nil))
}

func landingPage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "./login", 301)
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	t.ExecuteTemplate(w, "login.html", nil)
}

func passChange(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/admin/passwordChange/"):]
	var state ViewState
	state.Users = getAllUsers()
	state.ID, _ = strconv.Atoi(id)
	state.Config = config

	restrictPage(w, r, true, true)

	w.WriteHeader(200)
	t.ExecuteTemplate(w, "passwordChange.html", state)
}

func genLabel(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/label/p/"):]
	intID, _ := strconv.Atoi(id)
	intID++
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
	session := restrictPage(w, r, true, false)
	title := r.URL.Path[len("/home/"):]
	var state ViewState
	if len(title) != 0 {
		if len(title) >= len("users/new") && title[0:len("users/new")] == "users/new" {
			state.Users = getAllUsers()
			state.PageTitle = "New User"
			state.Page = "newUser"
			state.Session = session
			state.Config = config
			state.ID = 0
			w.WriteHeader(200)
			t.ExecuteTemplate(w, "root.html", state)
		} else if len(title) >= len("users") && title[0:len("users")] == "users" {
			state.Users = getAllUsers()
			state.PageTitle = "Users"
			state.Page = "users"
			state.Session = session
			state.Config = config
			state.ID = 0
			w.WriteHeader(200)
			t.ExecuteTemplate(w, "root.html", state)
		} else if len(title) >= len("user/") && title[0:len("user/")] == "user/" {
			state.Users = getAllUsers()
			state.PageTitle = "Users"
			state.Page = "user"
			state.Session = session
			state.Config = config
			state.ID, _ = strconv.Atoi(title[len("user/"):])
			w.WriteHeader(200)
			t.ExecuteTemplate(w, "root.html", state)
		} else if title[0:6] == "plants" {
			state.PageTitle = "Plants"
			state.Page = "plants"
			if session.IsAdmin {
				state.Plants = getAllPlants(-1)
			} else {
				state.Plants = getAllPlants(session.UserID)
			}
			state.Session = session
			state.Config = config
			state.ID = 0
			w.WriteHeader(200)
			t.ExecuteTemplate(w, "root.html", state)
		} else if title[0:len("plant/")] == "plant/" {
			state.PageTitle = "Plant"
			state.Page = "plant"
			if session.IsAdmin {
				state.Plants = getAllPlants(-1)
			} else {
				state.Plants = getAllPlants(session.UserID)
			}
			state.Session = session
			state.Config = config
			state.ID, _ = strconv.Atoi(title[0:len("plant/")])
			w.WriteHeader(200)
			t.ExecuteTemplate(w, "root.html", state)
		} else if len(title) >= len("varieties/new") && title[0:len("varieties/new")] == "varieties/new" {
			state.PageTitle = "New Variety"
			state.Page = "newVariety"
			if session.IsAdmin {
				state.Plants = getAllPlants(-1)
			} else {
				state.Plants = getAllPlants(session.UserID)
			}
			state.Session = session
			state.Config = config
			state.ID = 0
			w.WriteHeader(200)
			t.ExecuteTemplate(w, "root.html", state)
		} else if len(title) >= len("varieties/trefle") && title[0:len("varieties/trefle")] == "varieties/trefle" {
			state.PageTitle = "New Trefle Variety"
			state.Page = "newVarietyTrefle"
			if session.IsAdmin {
				state.Plants = getAllPlants(-1)
			} else {
				state.Plants = getAllPlants(session.UserID)
			}
			state.Session = session
			state.Config = config
			state.ID = 0
			w.WriteHeader(200)
			t.ExecuteTemplate(w, "root.html", state)
		} else if title[0:len("varieties")] == "varieties" {
			state.PageTitle = "Varieties"
			state.Page = "varieties"
			state.Varieties = getAllVarieties()
			state.Session = session
			state.Config = config
			state.ID = 0
			w.WriteHeader(200)
			t.ExecuteTemplate(w, "root.html", state)
		}
	} else {
		state.PageTitle = "Dashboard"
		state.Page = "dashboard"
		state.Session = session
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
	redir := r.Form.Get("redirect")

	var userObj, loginState = loginUser(user, pass)

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
	http.Redirect(w, r, "../../../home/user/"+userID, 302)
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
