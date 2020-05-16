package main

import (
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserID    int
	Username  string
	Hash      string
	LastLogin string
}

type UserList struct {
	Users []User
}

func getAllUsers() UserList {
	var stmt, err = db.Query("select * from `users`")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	var usrs UserList
	for stmt.Next() {
		var usr User
		err = stmt.Scan(&usr.UserID, &usr.Username, &usr.Hash, &usr.LastLogin)
		if err != nil {
			logMessage("Core", "user lookup error")
			log.Fatal(err)
		}
		usrs.Users = append(usrs.Users, usr)
	}
	err = stmt.Err()
	if err != nil {
		log.Fatal(err)
	}
	return usrs
}

func getUser(username string, password string) (usr User, found int) {

	var stmt, err = db.Query("select * from `users` where username = ? ", username)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	found = 0
	for stmt.Next() {
		err = stmt.Scan(&usr.UserID, &usr.Username, &usr.Hash, &usr.LastLogin)
		if err != nil {
			logMessage("Core", "user lookup error")
			log.Fatal(err)
		}
		var er = bcrypt.CompareHashAndPassword([]byte(usr.Hash), []byte(password))
		if er == nil {
			found = 1
			return
		}
	}
	err = stmt.Err()
	if err != nil {
		log.Fatal(err)
	}
	return
}

func loginUser(username string, password string) bool {
	var _, err = getUser(username, password)
	if err == 0 {
		var _, er = db.Exec("update users set hash = ? where username = ?", string(hashAndSalt([]byte(password))), username)
		if er != nil {
			logMessage("Core", "user update err")
			//log.Fatal(err)
		}
		return false
	} else {
		t := time.Now()
		var _, er = db.Exec("update users set lastLogin = "+t.String()+" where username = ?", username)
		if er != nil {
			logMessage("Core", "user update err")
			//log.Fatal(err)
		}
		return true
	}
}

func hashAndSalt(pwd []byte) string {

	// Use GenerateFromPassword to hash & salt pwd.
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}
