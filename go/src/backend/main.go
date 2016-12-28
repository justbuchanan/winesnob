package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"backend/apiai"
	"strings"

	"encoding/json"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type WineInfo struct {
	Name        string `json:"name"`
	Variety     string `json:"variety"`
	Description string `json:"description"`
	Year        int64  `json:"year"`
	Red         bool   `json:"red"`
	Available bool `json:"available"`

	Id          string `json:"id"`
}

var db *gorm.DB

func main() {
	// TODO: seed rng

	var err error
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

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

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/webhook", WebhookHandler).Methods("POST")

	// "api" routes
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/wine/{wineId}", WineHandler).Methods("GET")
	api.HandleFunc("/wine/{wineId}", WineDeleteHandler).Methods("DELETE")
	api.HandleFunc("/wines", WineCreateHandler).Methods("POST")
	api.HandleFunc("/wines", WinesIndexHandler).Methods("GET")
	api.HandleFunc("/wine/{wineId}", WineUpdateHandler).Methods("PUT")

	// serve angular frontend
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./dist/")))

	fmt.Println("Winesnob listening on port 8080")
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
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

func WineDescriptorLookup(descriptor map[string]interface{}) *WineInfo {
	var wines []WineInfo
	db.Find(&wines)	
	for _, wine := range wines {
		if wine.Variety == descriptor["variety"].(string) {
			fmt.Println("found it!")
			return &wine
		}
		fmt.Println("variety = ", wine.Variety)
	}

	return nil
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
				wineNames = append(wineNames, elem.Variety)
			}
		}

		resp.Speech = "We have " + JoinWordSeries(wineNames) + "."
	} else if intent == "wine.describe" {
		wineP := req.Result.Parameters["wine-descriptor"].(map[string]interface{})
		wine := WineDescriptorLookup(wineP)
		if wine != nil {
			resp.Speech = wine.Description
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

func WinesIndexHandler(w http.ResponseWriter, r *http.Request) {
	var wines []WineInfo
	db.Find(&wines)

	json.NewEncoder(w).Encode(wines)
}
