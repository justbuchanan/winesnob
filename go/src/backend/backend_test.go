package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	// "github.com/gorilla/sessions"
	"backend/apiai"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
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
	var err error

	err = InitConfigInfo("../../../cellar-config.json.example")
	if err != nil {
		t.Fatal(err)
	}

	db, _ = gorm.Open("sqlite3", ":memory:")
	defer db.Close()
	db.LogMode(false)

	// setup schema
	db.AutoMigrate(&WineInfo{})

	// run http server
	ts := httptest.NewServer(CreateHttpHandler())
	defer ts.Close()

	var res *http.Response

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
	req, _ := http.NewRequest("GET", ts.URL+"/api/wines", nil)
	ForceAuthenticate(req, "justbuchanan@gmail.com")
	fmt.Println("Force-authenticated as justbuchanan@gmail.com")

	res, err = http.Get(ts.URL + "/api/wines")
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(store)
	if res.StatusCode == http.StatusForbidden {
		// t.Fatal("Api should be accessible after user is authenticated")
		// TODO
	}

	fmt.Println("Trying empty action response")
	actionResponse := GetActionResponse(t, ts, &apiai.ActionRequest{})
	assert.Nil(t, actionResponse)

	RunTestWineDescriptorLookup(t)



	// request wine.describe(amarone)
	json_test := []byte(`
	{
	  "id": "1d6c7acd-745e-4bbd-88f8-d85dd74efa42",
	  "timestamp": "2017-01-01T03:17:25.553Z",
	  "result": {
	    "source": "agent",
	    "resolvedQuery": "tell me about the amarone",
	    "action": "",
	    "actionIncomplete": false,
	    "parameters": {
	      "wine-descriptor": "amarone"
	    },
	    "contexts": [],
	    "metadata": {
	      "intentId": "7b048d28-0bef-4b61-8e7a-c1cda95e01bc",
	      "webhookUsed": "true",
	      "webhookForSlotFillingUsed": "false",
	      "intentName": "wine.describe"
	    },
	    "fulfillment": {
	      "speech": "I'm sorry, I couldn't find a wine matching that description",
	      "source": "",
	      "displayText": "",
	      "messages": [
	        {
	          "type": 0,
	          "speech": "I'm sorry, I couldn't find a wine matching that description"
	        }
	      ],
	      "data": {}
	    },
	    "score": 1
	  },
	  "status": {
	    "code": 200,
	    "errorType": "success"
	  },
	  "sessionId": "e6a263ca-6ded-45df-bf76-3c960856e0bd"
	}`)

	var testReq apiai.ActionRequest
	err = json.Unmarshal(json_test, &testReq)
	if err != nil {
		t.Fatal(err)
	}

	testResp := GetActionResponse(t, ts, &testReq)
	assert.NotNil(t, testResp)
	assert.Equal(t, "amarone: Amarone description", testResp.Speech)
	fmt.Println("winner!", testResp.Speech)
}

func GetActionResponse(t *testing.T, ts *httptest.Server, req *apiai.ActionRequest) *apiai.ActionResponse {
	jsonValue, _ := json.Marshal(req)
	res, err := http.Post(ts.URL+"/webhook", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		t.Fatal(err)
		return nil
	}

	body, _ := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var apiResp apiai.ActionResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		fmt.Println(err)
		return nil
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

	// fmt.Println(session)
}

func RunTestWineDescriptorLookup(t *testing.T) {
	wines, err := ReadWinesFromFile("test-wines.json")
	if err != nil {
		t.Fatal(err)
	}

	// insert into database
	for _, wine := range wines {
		wine.Id = GenerateUniqueId()
		err = db.Create(&wine).Error
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}
	fmt.Println("Loaded test wines into db")

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
}
