package main

import (
    "net/http"
    "github.com/gorilla/mux"
    "encoding/json"
    "fmt"
    "log"
    "github.com/jinzhu/gorm"
)

func WineDeleteHandler(w http.ResponseWriter, r *http.Request) {
    if !EnsureLoggedIn(w, r) { return }

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
    if !EnsureLoggedIn(w, r) { return }

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
    if !EnsureLoggedIn(w, r) { return }

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
    if !EnsureLoggedIn(w, r) { return }

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
    if !EnsureLoggedIn(w, r) { return }

    // TODO: separate wine lists

    var wines []WineInfo
    db.Find(&wines)

    json.NewEncoder(w).Encode(wines)
}
