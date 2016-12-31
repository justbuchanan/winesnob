package main

import (
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

	"github.com/renstrom/fuzzysearch/fuzzy"
)

const SESSION_NAME = "cellar"

type WineInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Year        int64  `json:"year"`
	Red         bool   `json:"red"`
	Available   bool   `json:"available"`

	Id string `json:"id"`
}

type ServerConfigInfo struct {
	GoogleClientID string
	GoogleClientSecret string
	BaseURL string // example: https://cellar.justbuchanan.com:3000
	CookieSecret string
}

var db *gorm.DB
var store *sessions.CookieStore

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
	router.HandleFunc("/oauth2/logout", handleGoogleLogout)
	router.HandleFunc("/oauth2/google-callback", handleGoogleCallback)
	router.HandleFunc("/oauth2/login-status", LoginStatusHandler)

	// serve angular frontend
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./dist/")))

	loggedRouter := handlers.LoggingHandler(os.Stdout, router)

	return loggedRouter
}

func ReadConfigFile(filename string) (cfgRet *ServerConfigInfo, err error) {
	var file []byte
	file, err = ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg ServerConfigInfo

	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func main() {
	// TODO: seed rng

	loadSamples := flag.Bool("load-samples", false, "Load samples from wine-list.json")
	dbPath := flag.String("dbpath", "./wines.sqlite3db", "Path to sqlite3 database file")
	cfgPath := flag.String("config", "./cellar-config.json", "Path to server config file")

	port := flag.String("port", "8080", "listen on port")

	flag.Parse()

	var err error

	// load configuration
	var cfg *ServerConfigInfo
	cfg, err = ReadConfigFile(*cfgPath)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	store = sessions.NewCookieStore([]byte(cfg.CookieSecret))

	googleOauthConfig.RedirectURL = cfg.BaseURL + "/oauth2/google-callback"
	googleOauthConfig.ClientID = cfg.GoogleClientID
	googleOauthConfig.ClientID = cfg.GoogleClientSecret

	// sqlite3 database
	fmt.Printf("Connecting to database: %q\n", *dbPath)
	// TODO: error handling?
	db, _ = gorm.Open("sqlite3", *dbPath)
	defer db.Close()
	db.LogMode(true)

	// setup schema
	db.AutoMigrate(&WineInfo{})

	if *loadSamples == true {
		var wines []WineInfo
		wines, err = ReadWinesFromFile("wine-list.json")
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
	fmt.Println("Winesnob listening on port " + *port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+*port, loggedRouter))
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

func ReadWinesFromFile(filename string) (wines []WineInfo, err error) {
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

// does a fuzzy match against the all wine names and returns the top one if any
// match decently.
func WineDescriptorLookup(descriptor string) *WineInfo {
	var wines []WineInfo
	db.Find(&wines)

	const debug = false

	const WORST_ACCEPTABLE = 6
	var bestMatch WineInfo
	bestMatchR := WORST_ACCEPTABLE

	if debug {
		fmt.Println("Looking for", descriptor)
	}

	for _, wine := range wines {
		r := fuzzy.RankMatch(descriptor, wine.Name)
		if debug {
			fmt.Printf("  %d ", r)
			fmt.Println(wine.Name)
		}
		if r != -1 && r < bestMatchR {
			bestMatch = wine
			bestMatchR = r
		}
	}

	if bestMatchR < WORST_ACCEPTABLE {
		if debug {
			fmt.Println("Found:", bestMatch.Name)
		}
		return &bestMatch
	} else {
		return nil
	}
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

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
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
