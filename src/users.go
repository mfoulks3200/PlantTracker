package main

type User struct {
	UserID      int
	Username    string
	Permissions []Permission
}

type Permission struct {
	Path string
}

func init() {
	RegisterDatabaseCreationStatement(`CREATE TABLE "users" (
		"userID"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
		"username"	TEXT NOT NULL UNIQUE,
		"hash"	TEXT,
		"lastLogin"	TEXT
	)`)
}
