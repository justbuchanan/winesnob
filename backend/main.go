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
	Red         bool   `json:"red"`
	Available   bool   `json:"available"`

	Id string `json:"id"`
}

// see cellar-config.json.example for info
type ServerConfigInfo struct {
	GoogleClientID     string
	GoogleClientSecret string
	BaseURL            string
	CookieSecret       string
	ApiaiAuthUsername  string
	ApiaiAuthPassword  string
}

var store *sessions.CookieStore

type Env struct {
	db    *gorm.DB
	store *sessions.CookieStore
	// logger *log.logger
}

// these values are set via the server config file
var (
	APIAI_AUTH_USERNAME = "apiai"
	APIAI_AUTH_PASSWORD = "?"
)

func (env *Env) BasicAuthHandler(username string, password string, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()
		// if incorrect credentials, forbid access and return
		if !(user == APIAI_AUTH_USERNAME && pass == APIAI_AUTH_PASSWORD) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// if everything checks out, run next handler
		next(w, r)
	})
}

func (env *Env) CreateHttpHandler() http.Handler {
	router := mux.NewRouter().StrictSlash(true)

	// apiai webhook for "fulfillment"
	router.Handle("/webhook",
		env.BasicAuthHandler(APIAI_AUTH_USERNAME, APIAI_AUTH_PASSWORD,
			env.ApiaiWebhookHandler)).Methods("POST")

	// "api" routes
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/wine/{wineId}", env.WineHandler).Methods("GET")
	api.HandleFunc("/wine/{wineId}", env.WineDeleteHandler).Methods("DELETE")
	api.HandleFunc("/wines", env.WineCreateHandler).Methods("POST")
	api.HandleFunc("/wines", env.WinesIndexHandler).Methods("GET")
	api.HandleFunc("/wine/{wineId}", env.WineUpdateHandler).Methods("PUT")

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

func InitConfigInfo(filename string) error {
	// load configuration
	cfg, err := ReadConfigFile(filename)
	if err != nil {
		return err
	}

	// init cookie store
	store = sessions.NewCookieStore([]byte(cfg.CookieSecret))

	// global oauth config
	googleOauthConfig.RedirectURL = cfg.BaseURL + "/oauth2/google-callback"
	googleOauthConfig.ClientID = cfg.GoogleClientID
	googleOauthConfig.ClientSecret = cfg.GoogleClientSecret

	// http basic auth for apiai
	APIAI_AUTH_USERNAME = cfg.ApiaiAuthUsername
	APIAI_AUTH_PASSWORD = cfg.ApiaiAuthPassword

	if APIAI_AUTH_USERNAME == "" || APIAI_AUTH_PASSWORD == "" {
		log.Fatal("Apiai basic auth not set correctly in config file:", filename)
	}

	return nil
}

func LoadSamplesIntoDb(db *gorm.DB, filename string) error {
	wines, err := ReadWinesFromFile(filename)
	if err != nil {
		return err
	}

	// insert into database
	for _, wine := range wines {
		wine.Id = GenerateUniqueId(db)
		err = db.Create(&wine).Error
		if err != nil {
			return err
		}
	}

	fmt.Println("Loaded sample wines")

	return nil
}

func main() {
	// TODO: seed rng

	loadSamples := flag.Bool("load-samples", false, "Load samples from wine-list.json")
	dbPath := flag.String("dbpath", "./wines.sqlite3db", "Path to sqlite3 database file")
	cfgPath := flag.String("config", "./cellar-config.json", "Path to server config file")

	port := flag.String("port", "8080", "listen on port")

	flag.Parse()

	var err error

	err = InitConfigInfo(*cfgPath)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// sqlite3 database
	fmt.Printf("Connecting to database: %q\n", *dbPath)
	var db *gorm.DB
	db, err = gorm.Open("sqlite3", *dbPath)
	defer db.Close()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	db.LogMode(true)

	// setup schema
	db.AutoMigrate(&WineInfo{})

	if *loadSamples {
		err = LoadSamplesIntoDb(db, "wine-list.json")
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	env := &Env{}
	env.db = db
	env.store = store

	// log all requests
	loggedRouter := handlers.LoggingHandler(os.Stdout, env.CreateHttpHandler())
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

func GenerateUniqueId(db *gorm.DB) string {
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

// http://stackoverflow.com/questions/15323767/does-golang-have-if-x-in-construct-similar-to-python
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
