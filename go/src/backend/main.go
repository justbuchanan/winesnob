package main

import (
	"backend/apiai"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"encoding/json"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/renstrom/fuzzysearch/fuzzy"
)

type WineInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Year        int64  `json:"year"`
	Red         bool   `json:"red"`
	Available   bool   `json:"available"`

	Id string `json:"id"`
}

var db *gorm.DB
var store = sessions.NewCookieStore([]byte("something-very-secret")) // TODO: secret

func CreateHttpHandler() http.Handler {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/webhook", WebhookHandler).Methods("POST")

	// "api" routes
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/wine/{wineId}", WineHandler).Methods("GET")
	api.HandleFunc("/wine/{wineId}", WineDeleteHandler).Methods("DELETE")
	api.HandleFunc("/wines", WineCreateHandler).Methods("POST")
	api.HandleFunc("/wines", WinesIndexHandler).Methods("GET")
	api.HandleFunc("/wine/{wineId}", WineUpdateHandler).Methods("PUT")

    router.HandleFunc("/oauth2/login", handleGoogleLogin)
    router.HandleFunc("/oauth2/google-callback", handleGoogleCallback)

	// serve angular frontend
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./dist/")))
    
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)

	return loggedRouter
}

func main() {
	// TODO: seed rng

	loadSamples := flag.Bool("load-samples", false, "Load samples from wine-list.json")
	dbPath := flag.String("dbpath", "./wines.sqlite3db", "Path to sqlite3 database file")
	flag.Parse()

	// sqlite3 database
	fmt.Printf("Connecting to database: %q\n", *dbPath)
	// TODO: error handling?
	db, _ = gorm.Open("sqlite3", *dbPath)
	defer db.Close()
	db.LogMode(true)

	// setup schema
	db.AutoMigrate(&WineInfo{})


	if *loadSamples == true {
		wines, err := ReadWines("wine-list.json")
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		// insert into database
		for _, wine := range wines {
			wine.Id = GenerateUniqueId()
			err = db.Create(&wine).Error
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
		}

		fmt.Println("Loaded sample wines")
	}

	loggedRouter := CreateHttpHandler()
	fmt.Println("Winesnob listening on port 8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", loggedRouter))
}

// Similar to string Join(), but adds an "and" appropriately
func JoinWordSeries(items []string) string {
	if len(items) == 0 {
		return ""
	} else if len(items) == 1 {
		return items[0]
	} else {
		return strings.Join(items[:len(items)-1], ", ") + ", and " + items[len(items)-1]
	}
}

func ReadWines(filename string) (wines []WineInfo, err error) {
	// wines := make([]WineInfo, 4)
	var file []byte
	file, err = ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(file, &wines)
	if err != nil {
		return nil, err
	}

	return wines, nil
}

func WineDescriptorLookup(descriptor string) *WineInfo {
	var wines []WineInfo
	db.Find(&wines)

	var bestMatch *WineInfo
	bestMatchR := 0

	for _, wine := range wines {
		r := fuzzy.RankMatch(descriptor, wine.Name)
		if r > bestMatchR {
			bestMatch = &wine
			bestMatchR = r
		}
	}

	return bestMatch
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var req apiai.ActionRequest
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid request")
		return
	}

	var resp apiai.ActionResponse
	resp.Speech = "hello there!"
	resp.DisplayText = "what is this for?"

	intent := req.Result.Metadata.IntentName
	if intent == "wine.list" {
		color := req.Result.Parameters["wine-type"]
		var wineNames []string
		var wines []WineInfo
		db.Find(&wines)
		for _, elem := range wines {
			if (color == "" || (color == "red" == elem.Red)) && elem.Available {
				wineNames = append(wineNames, elem.Name)
			}
		}

		if len(wines) == 0 {
			resp.Speech = "Sad day... it looks like we're dry!"
		} else {
			resp.Speech = "We have " + JoinWordSeries(wineNames) + "."
		}

	} else if intent == "wine.describe" {
		wineDesc := req.Result.Parameters["wine-descriptor"].(string)
		wine := WineDescriptorLookup(wineDesc)
		if wine != nil {
			resp.Speech = wine.Name + ": " + wine.Description
		} else {
			resp.Speech = "I'm sorry, I couldn't find a wine matching that description"
		}
	} else if intent == "wine.pair" {
		food := req.Result.Parameters["food"].(string)
		resp.Speech = "I'd recommend the amarone, it goes very well with " + food
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Random generation borrowed from here:
// http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
func GenerateRandomId() string {
	const length = 4
	const letters = "abcdef0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GenerateUniqueId() string {
	for {
		id := GenerateRandomId()
		var count uint64
		err := db.Model(&WineInfo{}).Where("id = ?", id).Count(&count).Error
		if err != nil {
			log.Fatal(err)
			return ""
		}
		if count == 0 {
			return id
		}
	}
}

func WineDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	wineId := vars["wineId"]

	err := db.Delete(&WineInfo{Id: wineId}).Error
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, string(err.Error()))
		return
	}
	// TODO: set deleted status code
}

func WineCreateHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var wine WineInfo
	err := decoder.Decode(&wine)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid json")
		return
	}

	// assign a unique id
	wine.Id = GenerateUniqueId()

	// try to create a new wine in the db
	err = db.Create(&wine).Error
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, string(err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(wine)
}

func WineUpdateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	wineId := vars["wineId"]

	decoder := json.NewDecoder(r.Body)
	var wine WineInfo
	err := decoder.Decode(&wine)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid json")
		return
	}

	wine.Id = "" // clear wine id so it doesn't get set by the update
	err = db.Model(&wine).Where("id = ?", wineId).Updates(wine).Error
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, string(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(wine)
}

func WineHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	wineId := vars["wineId"]

	var wine WineInfo
	err := db.Where(&WineInfo{Id: wineId}).First(&wine).Error

	// 404 if no wine exists with that id
	if err == gorm.ErrRecordNotFound {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "No wine found for id %q\n", wineId)
		return
	}

	json.NewEncoder(w).Encode(wine)
}

func IsLoggedIn(r *http.Request) bool {
	session, err := store.Get(r, "session-name")
	if err != nil {
		// TODO: handle error
		log.Fatal(err)
		return false
	}

	email := session.Values["email"]
	return email != nil
}

func WinesIndexHandler(w http.ResponseWriter, r *http.Request) {
	if !IsLoggedIn(r) {
		http.Error(w, "need to authenticate", http.StatusForbidden)
	}

	// TODO: separate wine lists

	var wines []WineInfo
	db.Find(&wines)

	json.NewEncoder(w).Encode(wines)
}

var (
    googleOauthConfig = &oauth2.Config{
        RedirectURL:    "http://localhost:4200/oauth2/google-callback",
        ClientID: "1054965996082-b4rkamlpm0pou1v53h40kecds54d1h8p.apps.googleusercontent.com",
        ClientSecret: "lG7MbRpyc5joTUXPLyI9Ymft",
        Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile",
            "https://www.googleapis.com/auth/userinfo.email"},
        Endpoint:     google.Endpoint,
    }
// Some random string, random for each request
    oauthStateString = "random"
)

var ALLOWED_USERS = []string{"justbuchanan@gmail.com", "propinvestments@gmail.com"}

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
    url := googleOauthConfig.AuthCodeURL(oauthStateString)
    http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

type GoogleOauth2Result struct {
	ID string `json:"id"`
	Email string `json:"email"`
	VerifiedEmail bool `json:"verified_email"`
	Name string `json:"name"`
	GivenName string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Link string `json:"link"`
	Picture string `json:"picture"`
	Gender string `json:"gender"`
	Locale string `json:"locale"`
}

// http://stackoverflow.com/questions/15323767/does-golang-have-if-x-in-construct-similar-to-python
func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
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
		session, err := store.Get(r, "session-name")
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
