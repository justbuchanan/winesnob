package main

import (
	"encoding/json"
	"fmt"
	"github.com/justbuchanan/winesnob/backend/apiai"
	"log"
	"net/http"
)

func WineNotFoundResponse() *apiai.ActionResponse {
	return &apiai.ActionResponse{
		Speech: "I'm sorry, I couldn't find a wine matching that description"
	}
}

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
		return WineNotFoundResponse()
		// speech = "I'm sorry, I couldn't find a wine matching that description"
	}

	return &apiai.ActionResponse{
		Speech: speech,
	}
}

// TODO: actually implement this
func Intent_Wine_Pair(req apiai.ActionRequest) *apiai.ActionResponse {
	food := req.Result.Parameters["food"].(string)
	return &apiai.ActionResponse{
		Speech: "I'd recommend the amarone, it goes very well with " + food,
	}
}

// TODO: actually implement this
func Intent_Wine_Remove(req apiai.ActionRequest) *apiai.ActionResponse {
	// TODO
	return &apiai.ActionResponse{}
}

// func LookupWineDescriptorOrFail(descriptor string) *WineInfo {
// 	wine := WineDescriptorLookup(wineDesc)
// 	if wine == nil {

// 	}
	
// }

// TODO: actually implement this
func Intent_Wine_Add(req apiai.ActionRequest) *apiai.ActionResponse {
	// wineDesc := req.Result.Parameters["wine-descriptor"].(string)
	// wine := LookupWineDescriptorOrFail(wineDesc)
	// if wine == nil {
	// 	return
	// }

	// find closest match by descriptor
	wineDesc := req.Result.Parameters["wine-descriptor"].(string)

	if wine != nil {
		// TODO: mark this wine as available OR say that it's already available
	} else {
		// TODO: create new wine and say what happened
		// TODO: how to know if it's red or not?
	}

	// TODO
	return &apiai.ActionResponse{}
}

// TODO: actually implement this
func Intent_Wine_Query(req apiai.ActionRequest) *apiai.ActionResponse {
	// TODO
	return &apiai.ActionResponse{}
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
		"wine.list":               Intent_Wine_List,
		"wine.describe":           Intent_Wine_Describe,
		"wine.pair":               Intent_Wine_Pair,
		"wine.mark-unavailable":   Intent_Wine_Remove,
		"wine.mark-available":      Intent_Wine_Add,
		"wine.query-availability": Intent_Wine_Query,
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
