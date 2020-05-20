package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func FormatTitle(input string) string {
	words := strings.Fields(input)
	smallwords := " a an on the to "

	for index, word := range words {
		if strings.Contains(smallwords, " "+word+" ") {
			words[index] = word
		} else {
			words[index] = strings.Title(word)
		}
	}
	return strings.Join(words, " ")
}

func ReadTextFile(path string) string {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	// Convert []byte to string and print to screen
	text := string(content)
	return text
}

//Returns a textual representation of an Object as a string
func ObjectToString(a ...interface{}) string {
	return fmt.Sprintf("%v\n", a)
}

//Return true if there is a file at the path
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
