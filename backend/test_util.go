package main

import (
    "bytes"
    "encoding/json"
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/sqlite"
    "github.com/justbuchanan/winesnob/backend/apiai"
    "io/ioutil"
    "log"
    "net/http"
    "net/http/httptest"
    "os"
    "testing"
)


// configures, then initializes wine db with given file. Then executes each of
// the given functions and cleans up.
// Allows each test to be executed in a clean server state.
func WineContext(t *testing.T, f func(*testing.T, *httptest.Server)) {
    t.Log("Initializing context")

    err := InitConfigInfo("../cellar-config.json.example")
    if err != nil {
        t.Fatal(err)
    }

    db, _ = gorm.Open("sqlite3", ":memory:")
    defer db.Close()
    db.LogMode(false)

    // setup schema
    db.AutoMigrate(&WineInfo{})

    // run http server
    ts := httptest.NewServer(CreateHttpHandler())
    defer ts.Close()

    // run function
    f(t, ts)
}

func GetActionResponseFromJson(t *testing.T, ts *httptest.Server, jsonStr string) *apiai.ActionResponse {
    var testReq apiai.ActionRequest
    err := json.Unmarshal([]byte(jsonStr), &testReq)
    if err != nil {
        t.Fatal("parse error:", err)
    }

    return GetActionResponse(t, ts, &testReq)
}

func GetActionResponse(t *testing.T, ts *httptest.Server, req *apiai.ActionRequest) *apiai.ActionResponse {
    jsonValue, _ := json.Marshal(req)
    res, err := http.Post(ts.URL+"/webhook", "application/json", bytes.NewBuffer(jsonValue))
    if err != nil {
        t.Fatal(err)
        return nil
    }

    body, _ := ioutil.ReadAll(res.Body)
    if err != nil {
        t.Log(err)
        return nil
    }
    var apiResp apiai.ActionResponse
    err = json.Unmarshal(body, &apiResp)
    if err != nil {
        t.Log(err)
        return nil
    }

    return &apiResp
}

func ForceAuthenticate(ts *httptest.Server, email string) {
    req, _ := http.NewRequest("GET", ts.URL+"/api/wines", nil)
    session, err := store.Get(req, "session-name")
    if err != nil {
        log.Fatal(err)
        return
    }

    session.Values["email"] = email
    w := httptest.NewRecorder()
    session.Save(req, w)
}

func ClearDb() {
    db.Where("").Delete(&WineInfo{})
}


func LoadWinesFromJsonIntoDb(filename string) {
    wines, err := ReadWinesFromFile(filename)
    if err != nil {
        log.Fatal(err)
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
}
