package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

var (
	AppVersion  string
	AppBuildNum string
	AppCommit   string
	BuildDate   string
	BuildType   string
	Credits     []LibraryCredits
)

type Config struct {
	Webserver struct {
		Port     string `json:"port"`
		Hostname string `json:"hostname"`
	} `json:"webserver"`
	APIKeys struct {
		Trefle string `json:"trefleAPIKey"`
	} `json:"apis"`
	Sessions struct {
		Expiry string `json:"sessionExpiry"`
	} `json:"sessions"`
	Branding struct {
		Title string `json:"title"`
	} `json:"branding"`
	Hardware struct {
		LabelMaker struct {
			Manufacturer string `json:"Manufacturer"`
			LabelSize    string `json:"LabelSize"`
			Barcode      string `json:"Barcode"`
		} `json:"labelMaker"`
	} `json:"hardware"`
}

var config Config

func main() {
	initAppInfo()
	config = LoadConfiguration("./config.json")

	initDB()

	var wg sync.WaitGroup
	wg.Add(1)
	go startWebserver()
	wg.Wait()
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
