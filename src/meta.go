package main

type LibraryCredits struct {
	LibraryName    string
	LibraryAuthor  string
	LibraryVersion string
	LibraryLicense string
}

func initAppInfo() {
	logMessage("Core", "Starting Plant Tracker version "+AppVersion+" build "+AppBuildNum+" (commit "+AppCommit+" "+BuildDate+")")
}
