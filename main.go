package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/zmb3/spotify"
	"html/template"
	"log"
	"net/http"
	"path"
)

const redirectURI = "http://localhost:8080/callback"

var (
	auth = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadPrivate)

	//TODO: randomize it
	state = "state"
)

func main() {
	auth.SetAuthInfo("", "")

	router := mux.NewRouter()
	router.HandleFunc("/", handleHome)
	router.HandleFunc("/login", HandleLogin)
	router.HandleFunc("/callback", HandleCallback)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	layout := path.Join("template", "home", "layout.html")
	frontpage := path.Join("template", "home", "index.html")
	tmpl, err := template.ParseFiles(layout, frontpage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	url := auth.AuthURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleCallback(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != state {
		fmt.Println("State is not valid")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	token, err := auth.Token(state, r)
	if err != nil {
		fmt.Printf("Couldn't get token: %s\n", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	// use the token to get an authenticated client
	client := auth.NewClient(token)
	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "Login Completed %s!\n <img src='%s'>", user.DisplayName, user.Images[0].URL)
}
