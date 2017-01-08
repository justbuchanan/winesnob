package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var ALLOWED_USERS = []string{"justbuchanan@gmail.com", "propinvestments@gmail.com"}

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:4200/oauth2/google-callback",
		ClientID:     "1054965996082-b4rkamlpm0pou1v53h40kecds54d1h8p.apps.googleusercontent.com",
		ClientSecret: "lG7MbRpyc5joTUXPLyI9Ymft",
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint: google.Endpoint,
	}
	// Some random string, random for each request
	oauthStateString = "random"
)

type GoogleOauth2Result struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Link          string `json:"link"`
	Picture       string `json:"picture"`
	Gender        string `json:"gender"`
	Locale        string `json:"locale"`
}

// sends forbidden http response and returns false if the user isn't authenticated
func EnsureLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	session, err := store.Get(r, SESSION_NAME)
	if err != nil {
		// TODO: handle error
		log.Fatal(err)
		return false
	}

	email := session.Values["email"]
	if email != nil && stringInSlice(email.(string), ALLOWED_USERS) {
		return true
	}

	http.Error(w, "need to authenticate", http.StatusForbidden)
	return false
}

func LoginStatusHandler(w http.ResponseWriter, r *http.Request) {
	sess, err := store.Get(r, SESSION_NAME)
	if err != nil {
		log.Fatal(err)
		// TODO: set error http
		return
	}

	email := sess.Values["email"]

	js := map[string]interface{}{
		"email": email,
	}

	json.NewEncoder(w).Encode(js)
}

func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handleGoogleCallback")
	state := r.FormValue("state")
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Println("Code exchange failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	var result GoogleOauth2Result
	err = json.Unmarshal(contents, &result)
	fmt.Println("Got user: " + result.Email)
	if stringInSlice(result.Email, ALLOWED_USERS) {
		session, err := store.Get(r, SESSION_NAME)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session.Values["email"] = result.Email
		session.Save(r, w)

		fmt.Println("Allowed user!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleGoogleLogout(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, SESSION_NAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// clear session login data
	session.Values["email"] = nil
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
