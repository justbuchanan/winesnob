package main

import (
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"strings"
	"testing"
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
	WineContext(t, func(t *testing.T, ts *httptest.Server, env *Env) {
		env.LoadWinesFromJSONIntoDb("test-wines.json")
		// request wine.describe(amarone)
		testResp := GetActionResponseFromJSON(t, ts, RequestDescribeAmarone)
		if testResp == nil {
			t.Fatal("wine.describe(amarone) -> nil response")
		}
		assert.Equal(t, "amarone: Amarone description", testResp.Speech)
		t.Log("Response:", testResp.Speech)

		// clear db
		env.ClearDb()
		var count uint64
		env.db.Model(&WineInfo{}).Count(&count)
		assert.Equal(t, 0, int(count))

		env.LoadWinesFromJSONIntoDb("../wine-list.json")

		// same request as before, but against a different wine list
		testResp = GetActionResponseFromJSON(t, ts, RequestDescribeAmarone)
		assert.NotNil(t, testResp)
	})
}

func TestMarkUnavailable(t *testing.T) {
	WineContext(t, func(t *testing.T, ts *httptest.Server, env *Env) {
		env.LoadWinesFromJSONIntoDb("../wine-list.json")
		// check that it's available
		qResp := GetActionResponseFromJSON(t, ts, RequestAvailabilityStagsLeapMerlot)
		if qResp == nil {
			t.Fatal("nil response")
		}
		if !strings.HasPrefix(qResp.Speech, "Yes") {
			t.Fatal("Merlot should start out available")
		}

		// delete it
		GetActionResponseFromJSON(t, ts, RequestDeleteStagsLeapMerlot)

		// ensure that it's not available
		qResp = GetActionResponseFromJSON(t, ts, RequestAvailabilityStagsLeapMerlot)
		if !strings.HasPrefix(qResp.Speech, "No") {
			t.Fatal("Merlot should be gone after marking it unavailable")
		}
	})
}

func TestWineDescriptorLookup(t *testing.T) {
	WineContext(t, func(t *testing.T, ts *httptest.Server, env *Env) {
		env.LoadWinesFromJSONIntoDb("test-wines.json")
		t.Log("Loaded test wines into db")

		// exact match
		result := env.WineDescriptorLookup("chiraz")
		if assert.NotNil(t, result) {
			assert.Equal(t, "chiraz", result.Name)
		}

		// approximate match
		result = env.WineDescriptorLookup("chardonay") // missing an "n"
		if assert.NotNil(t, result) {
			assert.Equal(t, "chardonnay", result.Name)
		}

		result = env.WineDescriptorLookup("2013 amarone")
		if assert.NotNil(t, result) {
			assert.Equal(t, "2013 amarone", result.Name)
		}

		// bad match
		result = env.WineDescriptorLookup("bla bla bla")
		assert.Nil(t, result)
	})
}
