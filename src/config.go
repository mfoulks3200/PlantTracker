package main

import (
	"encoding/json"
	"fmt"

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
	fileText := ReadTextFile("config.json")
	json.Unmarshal([]byte(fileText), &Config)
}

func GetConfigurationPath(path string) string {
	val, _ := jsonpath.Read(Config, "$."+path)
	return fmt.Sprintf("%v", val)
}
