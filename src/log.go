package main

import (
	"log"
	"runtime"
	"strings"
	"time"
)

//Log a message to the console, and the inbuilt logging file with timestamps and calling information
func LogMessage(msg string) {
	log.SetFlags(0)

	pc, _, _, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()
	lastSlash := strings.LastIndexByte(funcName, '/')
	if lastSlash < 0 {
		lastSlash = 0
	}
	lastDot := strings.LastIndexByte(funcName[lastSlash:], '.') + lastSlash

	prefix := "[" + FormatTitle(funcName[lastDot+1:]) + "][" + FormatTitle(funcName[lastDot+1:]) + "]"

	t := time.Now()
	prefix += "[" + string(t.Format("1/2/6 3:4:5")) + "]"

	log.Print(prefix + " " + msg)
}
