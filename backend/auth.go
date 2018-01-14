package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
)

type BasicAuthCreds struct {
	Username, Password string
}

func (env *Env) BasicAuthHandler(username string, password string, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()
		// if incorrect credentials, forbid access and return
		if !(user == env.ApiaiCreds.Username && pass == env.ApiaiCreds.Password) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// if everything checks out, run next handler
		next(w, r)
	})
}

// check oauth2 status before forwarding to next handler
func (env *Env) OAuthGate(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if env.CheckLoggedIn(w, r) {
			next(w, r)
		} else {
			http.Error(w, "need to authenticate", http.StatusForbidden)
		}
	})
}

// generate a random oauth state token
// borrowed from: http://skarlso.github.io/2016/06/12/google-signin-with-go/
func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

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

func (env *Env) CheckLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	// fake auth - only for testing
	if env.authenticateEveryoneAs != "" {
		return true
	}

	session, err := env.store.Get(r, WinesnobSessionName)
	if err != nil {
		log.Fatal(err)
	}

	email := session.Values["email"]
	if email != nil && StringInSlice(email.(string), env.AllowedUsers) {
		return true
	}

	return false
}

func (env *Env) LoginStatusHandler(w http.ResponseWriter, r *http.Request) {
	var email string

	if env.authenticateEveryoneAs != "" {
		// mock auth for testing
		email = env.authenticateEveryoneAs
	} else {
		// check session cookie
		sess, err := env.store.Get(r, WinesnobSessionName)
		if err != nil {
			SendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		v, ok := sess.Values["email"]
		if ok {
			email = v.(string)
		} else {
			SendError(w, http.StatusForbidden, "Not logged in")
		}
	}

	js := map[string]interface{}{
		"email": email,
	}

	json.NewEncoder(w).Encode(js)
}

func (env *Env) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handleGoogleCallback")

	// get session
	session, err := env.store.Get(r, WinesnobSessionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// email := session.Values["email"]
	serverState := session.Values["state"]
	clientState := r.FormValue("state")
	if clientState != serverState {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", serverState, clientState)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := env.GoogleOauth2Config.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Println("Code exchange failed with:", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		fmt.Println("Bad response from google oauth2")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	var result GoogleOauth2Result
	err = json.Unmarshal(contents, &result)
	fmt.Println("Got user: " + result.Email)
	if StringInSlice(result.Email, env.AllowedUsers) {
		session.Values["email"] = result.Email
		session.Save(r, w)

		fmt.Println("Allowed user!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}

func (env *Env) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	session, err := env.store.Get(r, WinesnobSessionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	state := randToken()
	session.Values["state"] = state
	session.Save(r, w)

	url := env.GoogleOauth2Config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (env *Env) handleGoogleLogout(w http.ResponseWriter, r *http.Request) {
	session, err := env.store.Get(r, WinesnobSessionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// clear session login data
	session.Values["email"] = nil
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
