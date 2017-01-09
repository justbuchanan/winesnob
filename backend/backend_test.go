package main

import (
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
	"strings"
)


const RequestDescribeAmarone = `
	{
	  "result": {
	    "source": "agent",
	    "resolvedQuery": "tell me about the amarone",
	    "parameters": {
	      "wine-descriptor": "amarone"
	    },
	    "contexts": [],
	    "metadata": {
	      "intentName": "wine.describe"
	    }
	  }
	}`
const RequestAvailabilityStagsLeapMerlot = `
	{
	  "result": {
	    "parameters": {
	      "wine-descriptor": "Stags Leap Merlot"
	    },
	    "metadata": {
	      "intentName": "wine.query-availability"
	    }
	  }
	}`
const RequestDeleteStagsLeapMerlot = `
	{
		"result": {
			"parameters": {
				"wine-descriptor": "Stags Leap Merlot"
			},
			"metadata": {
				"intentName": "wine.mark-unavailable"
			}
		}
	}`


func TestDescribeWines(t *testing.T) {
	WineContext(t, func(t *testing.T, ts *httptest.Server) {
		LoadWinesFromJsonIntoDb("test-wines.json")
		// request wine.describe(amarone)
		testResp := GetActionResponseFromJson(t, ts, RequestDescribeAmarone)
		if testResp == nil {
			t.Fatal("wine.describe(amarone) -> nil response")
		}
		assert.Equal(t, "amarone: Amarone description", testResp.Speech)
		t.Log("winner!", testResp.Speech)

		// clear db
		ClearDb()
		var count uint64
		db.Model(&WineInfo{}).Count(&count)
		assert.Equal(t, 0, count)

		LoadWinesFromJsonIntoDb("../../../wine-list.json")

		// same request as before, but against a different wine list
		testResp = GetActionResponseFromJson(t, ts, RequestDescribeAmarone)
		assert.NotNil(t, testResp)
	})
}

func TestMarkUnavailable(t *testing.T) {
	WineContext(t, func(t *testing.T, ts *httptest.Server) {
		// check that it's available
		qResp := GetActionResponseFromJson(t, ts, RequestAvailabilityStagsLeapMerlot)
		if qResp == nil {
			t.Fatal("nil response")
		}
		if !strings.HasPrefix(qResp.Speech, "Yes") {
			t.Fatal("Merlot should start out available")
		}

		// delete it
		GetActionResponseFromJson(t, ts, RequestDeleteStagsLeapMerlot)

		// ensure that it's not available
		qResp = GetActionResponseFromJson(t, ts, RequestAvailabilityStagsLeapMerlot)
		if !strings.HasPrefix(qResp.Speech, "No") {
			t.Fatal("Merlot should be gone after marking it unavailable")
		}
	})
}

func TestWineDescriptorLookup(t *testing.T) {
	WineContext(t, func(t *testing.T, ts *httptest.Server) {
		LoadWinesFromJsonIntoDb("test-wines.json")
		t.Log("Loaded test wines into db")

		// exact match
		result := WineDescriptorLookup("chiraz")
		if assert.NotNil(t, result) {
			assert.Equal(t, "chiraz", result.Name)
		}

		// approximate match
		result = WineDescriptorLookup("chardonay") // missing an "n"
		if assert.NotNil(t, result) {
			assert.Equal(t, "chardonnay", result.Name)
		}

		result = WineDescriptorLookup("2013 amarone")
		if assert.NotNil(t, result) {
			assert.Equal(t, "2013 amarone", result.Name)
		}

		// bad match
		result = WineDescriptorLookup("bla bla bla")
		assert.Nil(t, result)
	})
}
