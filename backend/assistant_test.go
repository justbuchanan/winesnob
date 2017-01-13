package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

const RequestDescribeAmarone = `
    {
      "result": {
        "parameters": {
          "wine-descriptor": "amarone"
        },
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
const RequestListWines = `
    {
        "result": {
            "metadata": {
                "intentName": "wine.list"
            }
        }
    }`

func TestDescribeAmarone1(t *testing.T) {
	env, ts, cleanup := SetupTestServer(t)
	defer cleanup()

	LoadWinesFromFileIntoDb(env.db, "test-wines.json")
	testResp := env.GetActionResponseFromJSON(t, ts, RequestDescribeAmarone)
	if assert.NotNil(t, testResp) {
		assert.Equal(t, "amarone: Amarone description", testResp.Speech)
		t.Log("Response:", testResp.Speech)
	}
}

// same as above, but against a different wine list
func TestDescribeAmarone2(t *testing.T) {
	env, ts, cleanup := SetupTestServer(t)
	defer cleanup()

	LoadWinesFromFileIntoDb(env.db, "../wine-list.json")
	testResp := env.GetActionResponseFromJSON(t, ts, RequestDescribeAmarone)
	assert.NotNil(t, testResp)
	t.Log("Response:", testResp)
}

func TestMarkUnavailable(t *testing.T) {
	env, ts, cleanup := SetupTestServer(t)
	defer cleanup()

	LoadWinesFromFileIntoDb(env.db, "../wine-list.json")
	// check that it's available
	qResp := env.GetActionResponseFromJSON(t, ts, RequestAvailabilityStagsLeapMerlot)
	if qResp == nil {
		t.Fatal("nil response")
	}
	if !strings.HasPrefix(qResp.Speech, "Yes") {
		t.Fatal("Merlot should start out available")
	}

	// delete it
	env.GetActionResponseFromJSON(t, ts, RequestDeleteStagsLeapMerlot)

	// ensure that it's not available
	qResp = env.GetActionResponseFromJSON(t, ts, RequestAvailabilityStagsLeapMerlot)
	if !strings.HasPrefix(qResp.Speech, "No") {
		t.Fatal("Merlot should be gone after marking it unavailable")
	}
}

func TestWineDescriptorLookup(t *testing.T) {
	env, _, cleanup := SetupTestServer(t)
	defer cleanup()

	LoadWinesFromFileIntoDb(env.db, "test-wines.json")
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
}

func TestWebhookBlockedWhenNotLoggedIn(t *testing.T) {
	_, ts, cleanup := SetupTestServer(t)
	defer cleanup()

	res, err := http.Post(ts.URL+"/webhook", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusForbidden {
		t.Log("Response:", res)
		t.Fatal("Apiai webhook should require http basic auth")
	}
}

func TestEmptyWineList(t *testing.T) {
	env, ts, cleanup := SetupTestServer(t)
	defer cleanup()

	qResp := env.GetActionResponseFromJSON(t, ts, RequestListWines)
	if assert.NotNil(t, qResp) {
		assert.True(t, strings.HasPrefix(qResp.Speech, "Sad day"))
	}
}
