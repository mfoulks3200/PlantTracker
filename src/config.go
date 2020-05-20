package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yalp/jsonpath"
)

type Configuration struct {
	Catagories []ConfigurationCatagory
}

type ConfigurationCatagory struct {
	Catagories []ConfigurationCatagory
	Items      ConfigurationItem
}

type ConfigurationItem struct {
	Key   string
	Value string
}

var Config interface{}

func ReadConfigurationFile() {
	if !FileExists("config.json") {
		LogMessage("Cannot continue, no config.json file found")
		os.Exit(1)
	}
	fileText := ReadTextFile("config.json")
	json.Unmarshal([]byte(fileText), &Config)
}

func GetConfigurationPath(path string) string {
	val, _ := jsonpath.Read(Config, "$."+path)
	return fmt.Sprintf("%v", val)
}
