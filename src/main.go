package main

func main() {
	ReadConfigurationFile()

	if !FileExists("./Database.db") {
		LogMessage("Did not detect existing database, creating new Database.db")
		FirstTimeDBInit()
	}

}
