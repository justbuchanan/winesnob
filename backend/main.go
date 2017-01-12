package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
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

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const SESSION_NAME = "cellar"

type WineInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Red         bool   `json:"red"`
	Available   bool   `json:"available"`

	ID int `json:"id" sql:"AUTO_INCREMENT" gorm:"primary_key"`
}

// see cellar-config.json.example for info
type ServerConfigInfo struct {
	GoogleClientID     string
	GoogleClientSecret string
	BaseURL            string
	CookieSecret       string
	ApiaiAuthUsername  string
	ApiaiAuthPassword  string
	AllowedUsers       []string // whitelist of google accounts (ex: ['justin@gmail.com'])
}

// server context used by all handlers
type Env struct {
	db                 *gorm.DB
	store              *sessions.CookieStore
	ApiaiCreds         BasicAuthCreds
	GoogleOauth2Config oauth2.Config
	AllowedUsers       []string

	authenticate_everyone_as string // only use for testing
}

func (env *Env) CreateHTTPHandler() http.Handler {
	router := mux.NewRouter().StrictSlash(true)

	// apiai webhook for "fulfillment"
	router.Handle("/webhook",
		env.BasicAuthHandler(env.ApiaiCreds.Username, env.ApiaiCreds.Password,
			env.ApiaiWebhookHandler)).Methods("POST")

	// "api" routes
	api := router.PathPrefix("/api").Subrouter()
	api.Handle("/wine/{wineID}", env.OAuthGate(env.WineHandler)).Methods("GET")
	api.Handle("/wine/{wineID}", env.OAuthGate(env.WineDeleteHandler)).Methods("DELETE")
	api.Handle("/wines", env.OAuthGate(env.WineCreateHandler)).Methods("POST")
	api.Handle("/wines", env.OAuthGate(env.WinesIndexHandler)).Methods("GET")
	api.Handle("/wine/{wineID}", env.OAuthGate(env.WineUpdateHandler)).Methods("PUT")

	router.HandleFunc("/oauth2/login", env.handleGoogleLogin)
	router.HandleFunc("/oauth2/logout", env.handleGoogleLogout)
	router.HandleFunc("/oauth2/google-callback", env.handleGoogleCallback)
	router.HandleFunc("/oauth2/login-status", env.LoginStatusHandler)

	// serve angular frontend
	// note: it must first be built with `ng build`
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./dist/")))

	return router
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

func (env *Env) LoadConfigInfo(filename string) error {
	// load configuration
	cfg, err := ReadConfigFile(filename)
	if err != nil {
		return err
	}

	// init cookie store
	env.store = sessions.NewCookieStore([]byte(cfg.CookieSecret))

	// oauth2 config
	env.GoogleOauth2Config = oauth2.Config{
		RedirectURL:  cfg.BaseURL + "/oauth2/google-callback",
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint: google.Endpoint,
	}

	// google account whitelist
	env.AllowedUsers = cfg.AllowedUsers

	// http basic auth for apiai
	env.ApiaiCreds = BasicAuthCreds{
		Username: cfg.ApiaiAuthUsername,
		Password: cfg.ApiaiAuthPassword,
	}

	return nil
}

func LoadWinesFromFileIntoDb(db *gorm.DB, filename string) error {
	wines, err := ReadWinesFromFile(filename)
	if err != nil {
		return err
	}

	// insert into database
	for _, wine := range wines {
		if err = db.Create(&wine).Error; err != nil {
			return err
		}
	}

	fmt.Println("Loaded sample wines")

	return nil
}

func main() {
	// command-line flags
	loadSamples := flag.Bool("load-samples", false, "Load samples from wine-list.json")
	dbPath := flag.String("dbpath", "./wines.sqlite3db", "Path to sqlite3 database file")
	cfgPath := flag.String("config", "./cellar-config.json", "Path to server config file")
	port := flag.String("port", "8080", "listen on port")
	flag.Parse()

	var err error

	// sqlite3 database
	fmt.Printf("Connecting to database: %q\n", *dbPath)
	var db *gorm.DB
	db, err = gorm.Open("sqlite3", *dbPath)
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(true)

	// setup schema
	db.AutoMigrate(&WineInfo{})

	if *loadSamples {
		err = LoadWinesFromFileIntoDb(db, "wine-list.json")
		if err != nil {
			log.Fatal(err)
		}
	}

	env := &Env{
		db: db,
	}

	// sets auth and initializes jkcookie store
	err = env.LoadConfigInfo(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	// log all requests
	loggedRouter := handlers.LoggingHandler(os.Stdout, env.CreateHTTPHandler())
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
func (env *Env) WineDescriptorLookup(descriptor string) *WineInfo {
	descriptor = strings.ToLower(descriptor)
	var wines []WineInfo
	env.db.Find(&wines)

	const debug = false

	const WORST_ACCEPTABLE = 6
	var bestMatch WineInfo
	bestMatchR := WORST_ACCEPTABLE

	if debug {
		fmt.Println("Looking for", descriptor)
	}

	for _, wine := range wines {
		r := fuzzy.RankMatch(descriptor, strings.ToLower(wine.Name))
		if debug {
			fmt.Printf("  %d %s\n", r, wine.Name)
		}
		if r != -1 && r < bestMatchR {
			bestMatch, bestMatchR = wine, r
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
