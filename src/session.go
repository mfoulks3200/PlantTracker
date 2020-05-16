package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Session struct {
	UserID            int
	Username          string
	Token             string
	ExpireTimeSeconds int
	IsAdmin           bool
}

type SessionList struct {
	Sessions []Session
}

var ActiveSessions SessionList

func randToken() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func createSession(userID int) (session Session) {
	sessionToken := randToken()
	sessionTokenString := string(sessionToken[:])
	session.UserID = userID
	session.Token = sessionTokenString
	session.ExpireTimeSeconds, _ = strconv.Atoi(config.Sessions.Expiry)
	session.IsAdmin = true
	ActiveSessions.Sessions = append(ActiveSessions.Sessions, session)
	logMessage("core", "Started a session: "+sessionToken)
	return
}

func getSession(token string) (session Session, err int) {
	for i := 0; i < len(ActiveSessions.Sessions); i++ {
		if ActiveSessions.Sessions[i].Token == token {
			err = 0
			session = ActiveSessions.Sessions[i]
			return
		}
	}
	err = 1
	return
}

func registerSessionWithClient(w http.ResponseWriter, userID int) {
	session := createSession(userID)
	expiry, _ := strconv.Atoi(config.Sessions.Expiry)
	expire := ((time.Now()).AddDate(0, 0, expiry))
	cookie := http.Cookie{
		Name:    "authenticationSession",
		Value:   session.Token,
		Expires: expire,
	}
	http.SetCookie(w, &cookie)
	logMessage("core", "Created Session Cookie for Token "+session.Token+" to expire at "+expire.Format("1/2/06 3:4:5"))
}

func checkLogin(w http.ResponseWriter, r *http.Request) (session Session, valid bool) {
	// Read cookie
	cookie, err := r.Cookie("authenticationSession")
	if err != nil {
		valid = false
		return
	}
	session, er := getSession(cookie.Value)
	if er == 1 {
		valid = false
		return
	}
	valid = true
	return
}

func restrictPage(w http.ResponseWriter, r *http.Request, requireLoggedIn bool, requireAdmin bool) (session Session) {
	session, valid := checkLogin(w, r)
	if !valid && requireLoggedIn {
		http.Redirect(w, r, "http://"+config.Webserver.Hostname+":"+config.Webserver.Port+"/login?e=2&r="+url.QueryEscape(r.URL.Path), 302)
	} else if !session.IsAdmin && requireAdmin {
		http.Redirect(w, r, "http://"+config.Webserver.Hostname+":"+config.Webserver.Port+"/home/permissions", 302)
	}
	return
}
