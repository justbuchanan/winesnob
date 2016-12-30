package main

import (
    "testing"
    "net/http/httptest"
    "net/http"
    "log"
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/sqlite"
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
}
