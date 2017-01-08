package main

import (
	"github.com/justbuchanan/winesnob/backend/apiai"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func Intent_Wine_List(req apiai.ActionRequest) *apiai.ActionResponse {
	var resp apiai.ActionResponse

	color := req.Result.Parameters["wine-type"]
	var wineNames []string
	var wines []WineInfo
	db.Find(&wines)
	for _, elem := range wines {
		if (color == "" || (color == "red" == elem.Red)) && elem.Available {
			wineNames = append(wineNames, elem.Name)
		}
	}

	if len(wines) == 0 {
		resp.Speech = "Sad day... it looks like we're dry!"
	} else {
		resp.Speech = "We have " + JoinWordSeries(wineNames) + "."
	}

	return &resp
}

func Intent_Wine_Describe(req apiai.ActionRequest) *apiai.ActionResponse {
	wineDesc := req.Result.Parameters["wine-descriptor"].(string)
	wine := WineDescriptorLookup(wineDesc)
	var speech string
	if wine != nil {
		speech = wine.Name + ": " + wine.Description
	} else {
		speech = "I'm sorry, I couldn't find a wine matching that description"
	}

	return &apiai.ActionResponse{
		Speech: speech,
	}
}

func Intent_Wine_Pair(req apiai.ActionRequest) *apiai.ActionResponse {
	food := req.Result.Parameters["food"].(string)
	return &apiai.ActionResponse{
		Speech: "I'd recommend the amarone, it goes very well with " + food,
	}
}

func ApiaiWebhookHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var req apiai.ActionRequest
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid request")
		return
	}

	intentHandlers := map[string](func(apiai.ActionRequest) *apiai.ActionResponse){
		"wine.list":     Intent_Wine_List,
		"wine.describe": Intent_Wine_Describe,
		"wine.pair":     Intent_Wine_Pair,
	}

	intent := req.Result.Metadata.IntentName
	handler := intentHandlers[intent]
	if handler == nil {
		log.Println("No handler found for intent: '" + intent + "'")

		w.WriteHeader(http.StatusBadRequest)

		return
	}
	resp := handler(req)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
