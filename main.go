package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Config struct {
	Webserver struct {
		Port string `json:"port"`
	} `json:"webserver"`
	Branding struct {
		Title string `json:"title"`
	} `json:"branding"`
}

var config Config

func handler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadFile("templates/login.html")
	fmt.Fprint(w, string(body))
}

func main() {
	logMessage("Core", "Starting Plant Tracker")
	config = LoadConfiguration("config.json")

	logMessage("Core", "Started Webserver on port "+config.Webserver.Port)

	http.HandleFunc("/login", handler)
	log.Fatal(http.ListenAndServe(":"+config.Webserver.Port, nil))
}

func LoadConfiguration(file string) Config {
	logMessage("Core", "Loading config file from "+file)
	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

func logMessage(source string, msg string) {
	log.SetPrefix(source + " ")
	log.Print(msg)
}
