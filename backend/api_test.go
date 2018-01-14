package main

import (
	"bytes"
	"encoding/json"
	"github.com/justbuchanan/winesnob/backend/apiai"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strconv"
	"testing"
)

func TestJoinWordSeries(t *testing.T) {
	result := JoinWordSeries([]string{"red", "blue", "green"})
	expected := "red, blue, and green"
	assert.Equal(t, expected, result)
}

func TestStringInSlice(t *testing.T) {
	assert.True(t, StringInSlice("b", []string{"a", "b", "c"}))
	assert.False(t, StringInSlice("d", []string{"a", "b", "c"}))
}

func TestApiBlockedWhenNotLoggedIn(t *testing.T) {
	_, ts, cleanup := SetupTestServer(t)
	defer cleanup()

	// test authentication required
	res, err := http.Get(ts.URL + "/api/wines")
	if err != nil {
		t.Log(err)
	}
	if res.StatusCode != http.StatusForbidden {
		t.Fatal("Api should be blocked when not authenticated")
	}
}

func TestEmptyResponse(t *testing.T) {
	env, ts, cleanup := SetupTestServer(t)
	defer cleanup()

	actionResponse := env.GetActionResponse(t, ts, &apiai.ActionRequest{})
	assert.Nil(t, actionResponse)
}

func TestFakeAuthentication(t *testing.T) {
	env, ts, cleanup := SetupTestServer(t)
	defer cleanup()

	req, err := http.NewRequest("GET", ts.URL+"/api/wines", nil)
	if err != nil {
		t.Fatal(err)
	}
	env.authenticateEveryoneAs = "justbuchanan@gmail.com" // fake auth

	var res *http.Response
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode == http.StatusForbidden {
		t.Fatal("Api should be accessible after user is authenticated")
	}
}

func TestLoginStatus(t *testing.T) {
	env, ts, cleanup := SetupTestServer(t)
	defer cleanup()

	env.authenticateEveryoneAs = "drunk@cellar.edu"
	res, err := http.Get(ts.URL + "/oauth2/login-status")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var obj map[string]string
	err = json.NewDecoder(res.Body).Decode(&obj)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("obj: ", obj)
	assert.Equal(t, env.authenticateEveryoneAs, obj["email"])
}

func TestLoginStatusWhenUnauthenticated(t *testing.T) {
	_, ts, cleanup := SetupTestServer(t)
	defer cleanup()

	res, err := http.Get(ts.URL + "/oauth2/login-status")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusForbidden, res.StatusCode)
}

func TestCreate(t *testing.T) {
	env, ts, cleanup := SetupTestServer(t)
	defer cleanup()

	env.authenticateEveryoneAs = "someone"

	// send request to create a wine
	wineToCreate := WineInfo{
		Name:        "amarone 1",
		Description: "In the fields of Tuscany...",
		Available:   true,
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
	var wine WineInfo
	err = json.NewDecoder(res.Body).Decode(&wine)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("=> Created: ", wine)
	assert.Equal(t, wineToCreate.Available, wine.Available)
	assert.Equal(t, wineToCreate.Name, wine.Name)
	assert.Equal(t, wineToCreate.Description, wine.Description)

	t.Log("Update...")
	wine.Name = "merlot"
	b, err = json.Marshal(wine)
	var updateReq *http.Request
	updateReq, err = http.NewRequest("PUT", ts.URL+"/api/wine/"+strconv.Itoa(wine.ID), bytes.NewBuffer(b))
	res, err = http.DefaultClient.Do(updateReq)
	if err != nil {
		t.Fatal(err)
	}
	if assert.NotNil(t, res) {
		assert.Equal(t, http.StatusOK, res.StatusCode)
	}

	t.Log("Get")
	res, err = http.Get(ts.URL + "/api/wine/" + strconv.Itoa(wine.ID))
	if err != nil {
		t.Fatal(err)
	}

	var gottenWine *WineInfo
	err = json.NewDecoder(res.Body).Decode(&gottenWine)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, wine.Name, gottenWine.Name)

	t.Log("Delete")
	var deleteReq *http.Request
	deleteReq, err = http.NewRequest("DELETE", ts.URL+"/api/wine/"+strconv.Itoa(wine.ID), nil)
	res, err = http.DefaultClient.Do(deleteReq)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, res.StatusCode)

	t.Log("Get again")
	res, err = http.Get(ts.URL + "/api/wine/" + strconv.Itoa(wine.ID))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}
