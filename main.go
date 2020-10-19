package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
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
	store = sessions.NewCookieStore([]byte("mySession"))
	m     = map[string]dataUser{}
)

type dataUser struct {
	token  *oauth2.Token
	state  string
	client *spotify.PrivateUser
	musics []string
}

func main() {
	auth.SetAuthInfo("3f1e5d78deb5408aa01fb484acdae228", "f47df7a04a65404ab8b6972c8791aeaa")

	router := mux.NewRouter()
	router.HandleFunc("/", HandleIndex)
	router.HandleFunc("/login", HandleLogin)
	router.HandleFunc("/callback", HandleCallback)
	router.HandleFunc("/index", HandleHome)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	layout := path.Join("template", "index", "layout.html")
	frontpage := path.Join("template", "index", "index.html")
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

	limit := 10
	opt := &spotify.Options{
		Country:   nil,
		Limit:     &limit,
		Offset:    nil,
		Timerange: nil,
	}
	fullTPage, _ := client.CurrentUsersTopTracksOpt(opt)
	trackList := fullTPage.Tracks
	musicName := []string{
		" ",
	}
	for _, x := range trackList {
		musicName = append(musicName, x.Name)
	}

	sendData := dataUser{
		token:  token,
		state:  state,
		client: user,
		musics: musicName,
	}

	m[sendData.client.User.DisplayName] = sendData

	session, _ := store.Get(r, "mySession")

	session.Values["name"] = sendData.client.User.DisplayName
	session.Save(r, w)
	//fmt.Fprintf(w, "Login Completed %s!\n <img src='%s'>", user.DisplayName, user)
	http.Redirect(w, r, "/index", http.StatusSeeOther)
}

func HandleHome(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "mySession")

	name := session.Values["name"]
	fmt.Println(state)
	fmt.Println(name)

	str := fmt.Sprintf("%v", name)
	fmt.Println(str)

	actual := m[str]
	//update views
	layout := path.Join("template", "home", "layout.html")
	frontpage := path.Join("template", "home", "index.html")
	tmpl, err := template.ParseFiles(layout, frontpage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, actual); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	//
}
