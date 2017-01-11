package main

import (
	"github.com/justbuchanan/winesnob/backend/apiai"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJoinWordSeries(t *testing.T) {
	result := JoinWordSeries([]string{"red", "blue", "green"})
	expected := "red, blue, and green"
	assert.Equal(t, expected, result)
}

func TestBlockedWhenNotLoggedIn(t *testing.T) {
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
		actionResponse := GetActionResponse(t, ts, &apiai.ActionRequest{})
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
