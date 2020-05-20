package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DatabaseCreationStatements []string

//Create database
func FirstTimeDBInit() {
	db, err := sql.Open("sqlite3", "./Database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for i := 0; i < len(DatabaseCreationStatements); i++ {
		_, err = db.Exec(DatabaseCreationStatements[i])
		if err != nil {
			log.Printf("%q: %s\n", err, DatabaseCreationStatements[i])
			return
		}
	}
}

//Register function to initalize database element
func RegisterDatabaseCreationStatement(statement string) {
	DatabaseCreationStatements = append(DatabaseCreationStatements, statement)
}
