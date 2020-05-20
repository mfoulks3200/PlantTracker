package main

func main() {
	ReadConfigurationFile()
	if !FileExists("./PlantTracker.db") {
		LogMessage("Did not detect existing database, creating new PlantTracker.db")
		FirstTimeDBInit()
	}
}
