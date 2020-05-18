package main

import "strconv"

type AppInfo struct {
	AppVersion  string
	AppBuildNum int
	AppCommit   string
	BuildDate   string
	BuildType   string
	Credits     []LibraryCredits
}

type LibraryCredits struct {
	LibraryName    string
	LibraryAuthor  string
	LibraryVersion string
	LibraryLicense string
}

var globalAppInfo AppInfo

func initAppInfo() {
	globalAppInfo.AppVersion = "0.0.1"
	logMessage("Core", "Starting Plant Tracker version "+globalAppInfo.AppVersion+" build "+strconv.Itoa(globalAppInfo.AppBuildNum)+" (commit "+globalAppInfo.AppCommit+" "+globalAppInfo.BuildDate+")")
}
