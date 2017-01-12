package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"strconv"
)

func SendError(w http.ResponseWriter, statusCode int, text string) {
	log.Println("Error", text)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": text})
}

func (env *Env) WineDeleteHandler(w http.ResponseWriter, r *http.Request) {
	wineID, err := strconv.Atoi(mux.Vars(r)["wineID"])
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = env.db.Delete(&WineInfo{ID: wineID}).Error
	if err != nil {
		log.Println(err)
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}
}

func (env *Env) WineCreateHandler(w http.ResponseWriter, r *http.Request) {
	var wine WineInfo
	err := json.NewDecoder(r.Body).Decode(&wine)
	if err != nil {
		SendError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// assign a unique id
	// wine.ID = GenerateUniqueID(env.db)

	// try to create a new wine in the db
	err = env.db.Create(&wine).Error
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(wine)
}

func (env *Env) WineUpdateHandler(w http.ResponseWriter, r *http.Request) {
	wineID, err := strconv.Atoi(mux.Vars(r)["wineID"])
	if err != nil {
		log.Println("at Update, wineID = ", mux.Vars(r)["wineID"])
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// TODO: ensure wineID already exists?

	var wine WineInfo
	err = json.NewDecoder(r.Body).Decode(&wine)
	if err != nil {
		SendError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	wine.ID = wineID

	err = env.db.Save(wine).Error
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	json.NewEncoder(w).Encode(wine)
}

func (env *Env) WineHandler(w http.ResponseWriter, r *http.Request) {
	wineID, err := strconv.Atoi(mux.Vars(r)["wineID"])
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	var wine WineInfo
	err = env.db.Where(&WineInfo{ID: wineID}).First(&wine).Error

	// 404 if no wine exists with that id
	if err == gorm.ErrRecordNotFound {
		SendError(w, http.StatusNotFound, "No wine found for id "+strconv.Itoa(wineID))
		return
	}

	json.NewEncoder(w).Encode(wine)
}

func (env *Env) WinesIndexHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: separate wine lists

	var wines []WineInfo
	env.db.Find(&wines)

	json.NewEncoder(w).Encode(wines)
}
