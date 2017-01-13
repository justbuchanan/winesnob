package main

import (
	"bytes"
	"encoding/json"
	"github.com/justbuchanan/winesnob/backend/apiai"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Note: need to defer ts.Close() and env.db.Close()
func SetupTestServer(t *testing.T) (*Env, *httptest.Server, func()) {
	t.Log("Initializing context")

	// setup Env and db
	env := &Env{}
	env.InitDB(":memory:")
	env.db.LogMode(false)

	err := env.LoadConfigInfo("../cellar-config.json.example")
	if err != nil {
		t.Fatal(err)
	}

	// run http server
	ts := httptest.NewServer(env.CreateHTTPHandler())

	cleanup := func() {
		ts.Close()
		env.db.Close()
	}

	return env, ts, cleanup
}

func (env *Env) GetActionResponseFromJSON(t *testing.T, ts *httptest.Server, jsonStr string) *apiai.ActionResponse {
	var testReq apiai.ActionRequest
	err := json.Unmarshal([]byte(jsonStr), &testReq)
	if err != nil {
		t.Fatal("parse error:", err)
	}

	return env.GetActionResponse(t, ts, &testReq)
}

func (env *Env) GetActionResponse(t *testing.T, ts *httptest.Server, req *apiai.ActionRequest) *apiai.ActionResponse {
	// build POST request with apiai request
	jsonValue, _ := json.Marshal(req)
	httpReq, err := http.NewRequest("POST", ts.URL+"/webhook", bytes.NewBuffer(jsonValue))
	if err != nil {
		t.Fatal(err)
		return nil
	}
	// use basic auth to authenticate
	httpReq.SetBasicAuth(env.ApiaiCreds.Username, env.ApiaiCreds.Password)

	// do it!
	var res *http.Response
	res, err = http.DefaultClient.Do(httpReq)
	if err != nil {
		t.Fatal(err)
		return nil
	}

	// parse response body json
	var apiResp apiai.ActionResponse
	err = json.NewDecoder(res.Body).Decode(&apiResp)
	if err != nil {
		t.Log(err)
		return nil
	}

	return &apiResp
}
