package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	// "time"
	// "apiai"
	"backend/apiai"

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
}

type Part struct {
	Id          string `json:"id"`
	Brief       string `json:"brief"`
	Description string `json:"description"`
	Quantity    uint64 `json:"quantity"`
}

var allWines []WineInfo
var db *gorm.DB

func main() {
	// TODO: seed rng

	var err error
	allWines, err = ReadWines("wine-list.json")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	dbPath := flag.String("dbpath", "./parts.sqlite3db", "Path to sqlite3 database file")
	flag.Parse()

	// sqlite3 database
	fmt.Printf("Connecting to database: %q\n", *dbPath)
	// TODO: error handling?
	db, _ = gorm.Open("sqlite3", *dbPath)
	defer db.Close()
	db.LogMode(true)

	// setup schema
	db.AutoMigrate(&Part{})

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/webhook", WebhookHandler).Methods("POST")

	// parts "api" routes
	// api := router.PathPrefix("/api").Subrouter()
	// api.HandleFunc("/part/{partId}", PartHandler).Methods("GET")
	// api.HandleFunc("/part/{partId}", PartDeleteHandler).Methods("DELETE")
	// api.HandleFunc("/part/{partId}/label", PartLabelHandler).Methods("GET")
	// api.HandleFunc("/parts", PartCreateHandler).Methods("POST")
	// api.HandleFunc("/parts", PartsIndexHandler).Methods("GET")
	// api.HandleFunc("/part/{partId}", PartUpdateHandler).Methods("PUT")

	// serve angular frontend
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./dist/")))

	fmt.Println("Winesnob listening on port 8080")
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", loggedRouter))
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
	for _, wine := range allWines {
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
		for _, elem := range allWines {
			if color == "" || (color == "red" == elem.Red) {
				wineNames = append(wineNames, elem.Variety)
			}
		}
		// if color == "" {
		// 	resp.Speech = "We have an italian amarone, a cabernet, and a chardonnay"
		// } else if color == "red" {
		// 	resp.Speech = "listing only red wines"
		// } else if color == "white" {
		// 	resp.Speech = "listing only white wines"
		// } else {
		// 	resp.Speech = "Unknown wine type"
		// }

		resp.Speech = "We have "
		for _, variety := range wineNames {
			resp.Speech += variety + ", "
		}
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
		err := db.Model(&Part{}).Where("id = ?", id).Count(&count).Error
		if err != nil {
			log.Fatal(err)
			return ""
		}
		if count == 0 {
			return id
		}
	}
}

func PartDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	partId := vars["partId"]

	err := db.Delete(&Part{Id: partId}).Error
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, string(err.Error()))
		return
	}
	// TODO: set deleted status code
}

func PartCreateHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var part Part
	err := decoder.Decode(&part)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid json")
		return
	}

	// assign a unique id
	part.Id = GenerateUniqueId()

	// try to create a new part in the db
	err = db.Create(&part).Error
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, string(err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(part)
}

func PartUpdateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	partId := vars["partId"]

	decoder := json.NewDecoder(r.Body)
	var part Part
	err := decoder.Decode(&part)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid json")
		return
	}

	part.Id = "" // clear part id so it doesn't get set by the update
	err = db.Model(&part).Where("id = ?", partId).Updates(part).Error
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, string(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(part)
}

func PartHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	partId := vars["partId"]

	var part Part
	err := db.Where(&Part{Id: partId}).First(&part).Error

	// 404 if no part exists with that id
	if err == gorm.ErrRecordNotFound {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "No part found for id %q\n", partId)
		return
	}

	json.NewEncoder(w).Encode(part)
}

func PartsIndexHandler(w http.ResponseWriter, r *http.Request) {
	var parts []Part
	db.Find(&parts)

	json.NewEncoder(w).Encode(parts)
}

func PartLabelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	partId := vars["partId"]

	// lookup part
	var part Part
	err := db.Where(&Part{Id: partId}).First(&part).Error

	// 404 if no part exists with that id
	if err == gorm.ErrRecordNotFound {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "No part found for id %q\n", partId)
		return
	}

	// create tmp dir to write label into
	var tmpdir string
	tmpdir, err = ioutil.TempDir("/tmp", "inventory")
	if err != nil {
		log.Fatal(err)
	}

	filename := part.Id + "-label.pdf"
	outpath := tmpdir + "/" + part.Id

	// generate label using python script
	dir, _ := os.Getwd()
	fmt.Println(dir)
	cmd := exec.Command(dir+"/dymo-labelgen/main.py",
		part.Brief,
		"https://inventory.justbuchanan.com/part/"+part.Id,
		"--bbox",
		"--size=small",
		"--output="+outpath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf(string(out.Bytes()))

	// read pdf file
	pdfdata, err := os.Open(outpath)
	if err != nil {
		log.Fatal(err)
	}
	defer pdfdata.Close()

	// get file size
	var fi os.FileInfo
	fi, err = pdfdata.Stat()
	if err != nil {
		log.Fatal(err)
	}
	pdfSize := fi.Size()

	// set header info
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Length", strconv.FormatInt(pdfSize, 10))

	// write pdf to http response
	_, err = io.Copy(w, pdfdata)
	if err != nil {
		log.Fatal(err)
	}
}
