package main

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

func printAppInfo() {
	logMessage("Core", "Starting Plant Tracker version "+globalAppInfo.AppVersion+" build "+globalAppInfo.AppBuildNum+" (commit "+globalAppInfo.AppCommit+")")
}
