package main

import (
	"github.com/justbuchanan/winesnob/backend/apiai"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"bytes"
	"encoding/json"
)

func TestJoinWordSeries(t *testing.T) {
	result := JoinWordSeries([]string{"red", "blue", "green"})
	expected := "red, blue, and green"
	assert.Equal(t, expected, result)
}

func TestApiBlockedWhenNotLoggedIn(t *testing.T) {
	WineContext(t, func(t *testing.T, ts *httptest.Server, env *Env) {
		// test authentication required
		res, err := http.Get(ts.URL + "/api/wines")
		if err != nil {
			t.Log(err)
		}
		if res.StatusCode != http.StatusForbidden {
			t.Fatal("Api should be blocked when not authenticated")
		}
	})
}

func TestEmptyResponse(t *testing.T) {
	WineContext(t, func(t *testing.T, ts *httptest.Server, env *Env) {
		actionResponse := env.GetActionResponse(t, ts, &apiai.ActionRequest{})
		assert.Nil(t, actionResponse)
	})
}

func TestFakeAuthentication(t *testing.T) {
	WineContext(t, func(t *testing.T, ts *httptest.Server, env *Env) {
		req, err := http.NewRequest("GET", ts.URL+"/api/wines", nil)
		if err != nil {
			t.Fatal(err)
		}
		env.authenticate_everyone_as = "justbuchanan@gmail.com" // fake auth

		var res *http.Response
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode == http.StatusForbidden {
			t.Fatal("Api should be accessible after user is authenticated")
		}
	})
}

func TestCreate(t *testing.T) {
	WineContext(t, func(t *testing.T, ts *httptest.Server, env *Env) {
		env.authenticate_everyone_as = "someone"

		// send request to create a wine
		wineToCreate := WineInfo{
			Name: "amarone 1",
			Description: "In the fields of Tuscany...",
			Available: true,
		}
		b, err := json.Marshal(wineToCreate)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("Creating wine: ", wineToCreate)
		var res *http.Response
		res, err = http.Post(ts.URL+"/api/wines", "application/json", bytes.NewBuffer(b))
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, http.StatusCreated, res.StatusCode)

		// decode response and make sure it's the same as we sent
		decoder := json.NewDecoder(res.Body)
		var wine WineInfo
		err = decoder.Decode(&wine)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("=> Created: ", wine)
		assert.Equal(t, wineToCreate.Available, wine.Available)
		assert.Equal(t, wineToCreate.Name, wine.Name)
		assert.Equal(t, wineToCreate.Description, wine.Description)
	})
}
