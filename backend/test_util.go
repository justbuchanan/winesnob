package main

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/sessions"
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
func WineContext(t *testing.T, f func(*testing.T, *httptest.Server, *Env)) {
	t.Log("Initializing context")

	err := InitConfigInfo("../cellar-config.json.example")
	if err != nil {
		t.Fatal(err)
	}

	db, _ := gorm.Open("sqlite3", ":memory:")
	defer db.Close()
	db.LogMode(false)

	// setup schema
	db.AutoMigrate(&WineInfo{})

	env := &Env{
		db:    db,
		store: sessions.NewCookieStore([]byte("wahphuR0eethoo8R")),
	}

	// run http server
	ts := httptest.NewServer(env.CreateHTTPHandler())
	defer ts.Close()

	// run function
	f(t, ts, env)
}

func GetActionResponseFromJSON(t *testing.T, ts *httptest.Server, jsonStr string) *apiai.ActionResponse {
	var testReq apiai.ActionRequest
	err := json.Unmarshal([]byte(jsonStr), &testReq)
	if err != nil {
		t.Fatal("parse error:", err)
	}

	return GetActionResponse(t, ts, &testReq)
}

func GetActionResponse(t *testing.T, ts *httptest.Server, req *apiai.ActionRequest) *apiai.ActionResponse {
	// build POST request with apiai request
	jsonValue, _ := json.Marshal(req)
	httpReq, err := http.NewRequest("POST", ts.URL+"/webhook", bytes.NewBuffer(jsonValue))
	if err != nil {
		t.Fatal(err)
		return nil
	}
	// use basic auth to authenticate
	httpReq.SetBasicAuth(APIAI_AUTH_USERNAME, APIAI_AUTH_PASSWORD)

	// do it!
	var res *http.Response
	res, err = http.DefaultClient.Do(httpReq)
	if err != nil {
		t.Fatal(err)
		return nil
	}

	// parse response body json
	body, _ := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Log(err)
		return nil
	}
	t.Log("body: ", string(body))
	var apiResp apiai.ActionResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		t.Log(err)
		return nil
	}

	return &apiResp
}

func (env *Env) ClearDb() {
	env.db.Where("").Delete(&WineInfo{})
}

func (env *Env) LoadWinesFromJSONIntoDb(filename string) {
	wines, err := ReadWinesFromFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	// insert into database
	for _, wine := range wines {
		wine.ID = GenerateUniqueID(env.db)
		err = env.db.Create(&wine).Error
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}
}
