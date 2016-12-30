package main

import (
    "testing"
    "net/http/httptest"
    "net/http"
    "log"
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/sqlite"
    // "github.com/gorilla/sessions"
    "fmt"
    "bytes"
    "encoding/json"
    "backend/apiai"
    "io/ioutil"
)

func TestJoinWordSeries(t *testing.T) {
    items := []string{"red", "blue", "green"}
    result := JoinWordSeries(items)
    expected := "red, blue, and green"
    if result != expected {
        t.Error("result != expected")
    }
}

func TestApi(t *testing.T) {
    db, _ = gorm.Open("sqlite3", ":memory:")
    defer db.Close()
    db.LogMode(true)

    // setup schema
    db.AutoMigrate(&WineInfo{})

    // run http server
    ts := httptest.NewServer(CreateHttpHandler())
    defer ts.Close()

    var res *http.Response
    var err error

    // test authentication required
    res, err = http.Get(ts.URL + "/api/wines")
    if err != nil {
        log.Fatal(err)
    }
    if res.StatusCode != http.StatusForbidden {
        t.Fatal("Api should be blocked when not authenticated")
    }

    // var sess *Session
    // sess, err = store.Getjk
    req, _ := http.NewRequest("GET", ts.URL + "/api/wines", nil)
    ForceAuthenticate(req, "justbuchanan@gmail.com")
    fmt.Println("Force-authenticated as justbuchanan@gmail.com")


    res, err = http.Get(ts.URL + "/api/wines")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(store)
    if res.StatusCode == http.StatusForbidden {
        // t.Fatal("Api should be accessible after user is authenticated")
        // TODO
    }

    actionResponse := GetActionResponse(t, ts, &apiai.ActionRequest{})
    fmt.Println(actionResponse)
}

func GetActionResponse(t *testing.T, ts *httptest.Server, req *apiai.ActionRequest) *apiai.ActionResponse {
    jsonValue, _ := json.Marshal(req)
    res, err := http.Post(ts.URL + "/webhook", "application/json", bytes.NewBuffer(jsonValue))
    if err != nil {
        t.Fatal(err)
        return nil
    }

    var apiResp apiai.ActionResponse
    body, _ := ioutil.ReadAll(res.Body)
    if err != nil {
        t.Fatal(err)
    }
    err = json.Unmarshal(body, &apiResp)
    if err != nil {
        t.Fatal(err)
    }

    return &apiResp
}


func ForceAuthenticate(req *http.Request, email string) {
    session, err := store.Get(req, "session-name")
    if err != nil {
        log.Fatal(err)
        return
    }

    session.Values["email"] = email
    w := httptest.NewRecorder()
    session.Save(req, w)

    fmt.Println(session)
}
