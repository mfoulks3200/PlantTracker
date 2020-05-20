package main

import "sync"

func main() {
	ReadConfigurationFile()

	if !FileExists("./Database.db") {
		LogMessage("Did not detect existing database, creating new Database.db")
		FirstTimeDBInit()
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go RunWebserver()
	wg.Wait()

}
