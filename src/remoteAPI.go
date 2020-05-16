package main

import (
	"bytes"
	"fmt"
	"net/http"
)

func trefleQuery(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Path[len("/api/trefle/query/"):]
	w.WriteHeader(200)
	resp := getWebpage("https://trefle.io/api/plants/?token=" + config.APIKeys.Trefle + "&q=" + query)
	fmt.Fprint(w, resp)
}

func treflePlantByID(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Path[len("/api/trefle/plant/id/"):]
	w.WriteHeader(200)
	resp := getWebpage("https://trefle.io/api/plants/" + query + "?token=" + config.APIKeys.Trefle)
	fmt.Fprint(w, resp)
}

func getWebpage(url string) (body string) {
	logMessage("core", "Remote API Request: "+url)
	resp, err := http.Get(url)
	// handle the error if there is one
	if err != nil {
		panic(err)
	}
	// do this now so it won't be forgotten
	defer resp.Body.Close()
	// reads html as a slice of bytes
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body = buf.String()
	return
}
