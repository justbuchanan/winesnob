package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
)

func (env *Env) WineDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if !env.EnsureLoggedIn(w, r) {
		return
	}

	vars := mux.Vars(r)
	wineID := vars["wineId"]

	err := env.db.Delete(&WineInfo{ID: wineID}).Error
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, string(err.Error()))
		return
	}
	// TODO: set deleted status code
}

func (env *Env) WineCreateHandler(w http.ResponseWriter, r *http.Request) {
	if !env.EnsureLoggedIn(w, r) {
		return
	}

	decoder := json.NewDecoder(r.Body)
	var wine WineInfo
	err := decoder.Decode(&wine)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid json")
		return
	}

	// assign a unique id
	wine.ID = GenerateUniqueID(env.db)

	// try to create a new wine in the db
	err = env.db.Create(&wine).Error
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, string(err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(wine)
}

func (env *Env) WineUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if !env.EnsureLoggedIn(w, r) {
		return
	}

	vars := mux.Vars(r)
	wineID := vars["wineId"]

	// TODO: ensure wineID already exists?

	decoder := json.NewDecoder(r.Body)
	var wine WineInfo
	err := decoder.Decode(&wine)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid json")
		return
	}

	wine.ID = wineID

	err = env.db.Save(wine).Error
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, string(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(wine)
}

func (env *Env) WineHandler(w http.ResponseWriter, r *http.Request) {
	if !env.EnsureLoggedIn(w, r) {
		return
	}

	vars := mux.Vars(r)
	wineID := vars["wineId"]

	var wine WineInfo
	err := env.db.Where(&WineInfo{ID: wineID}).First(&wine).Error

	// 404 if no wine exists with that id
	if err == gorm.ErrRecordNotFound {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "No wine found for id %q\n", wineID)
		return
	}

	json.NewEncoder(w).Encode(wine)
}

func (env *Env) WinesIndexHandler(w http.ResponseWriter, r *http.Request) {
	if !env.EnsureLoggedIn(w, r) {
		return
	}

	// TODO: separate wine lists

	var wines []WineInfo
	env.db.Find(&wines)

	json.NewEncoder(w).Encode(wines)
}
