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
		Speech: "I'm sorry, I couldn't find a wine matching that description",
	}
}

func (env *Env) Intent_Wine_List(req apiai.ActionRequest) *apiai.ActionResponse {
	var resp apiai.ActionResponse

	color := req.Result.Parameters["wine-type"]
	var wineNames []string
	var wines []WineInfo
	env.db.Find(&wines)
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

func (env *Env) Intent_Wine_Describe(req apiai.ActionRequest) *apiai.ActionResponse {
	wineDesc := req.Result.Parameters["wine-descriptor"].(string)
	wine := env.WineDescriptorLookup(wineDesc)

	if wine == nil {
		return WineNotFoundResponse()
	}

	return &apiai.ActionResponse{
		Speech: wine.Name + ": " + wine.Description,
	}
}

// TODO: actually implement this
func (env *Env) Intent_Wine_Pair(req apiai.ActionRequest) *apiai.ActionResponse {
	food := req.Result.Parameters["food"].(string)
	return &apiai.ActionResponse{
		Speech: "I'd recommend the amarone, it goes very well with " + food,
	}
}

func (env *Env) Intent_Wine_Remove(req apiai.ActionRequest) *apiai.ActionResponse {
	wineDesc := req.Result.Parameters["wine-descriptor"].(string)
	wine := env.WineDescriptorLookup(wineDesc)

	if wine == nil {
		return WineNotFoundResponse()
	}

	if !wine.Available {
		return &apiai.ActionResponse{
			Speech: "Yes, I know there's no " + wine.Name + " available",
		}
	}

	// mark unavailable
	wine.Available = false
	err := env.db.Model(&wine).Where("id = ?", wine.Id).Updates(wine).Error
	if err != nil {
		return &apiai.ActionResponse{
			Speech: "Error", // TODO
		}
	}

	return &apiai.ActionResponse{
		Speech: "Noted, there's no more " + wine.Name + " left",
	}
}

func (env *Env) Intent_Wine_Add(req apiai.ActionRequest) *apiai.ActionResponse {
	wineDesc := req.Result.Parameters["wine-descriptor"].(string)
	wine := env.WineDescriptorLookup(wineDesc)

	if wine == nil {
		return WineNotFoundResponse()
	}

	if wine.Available {
		return &apiai.ActionResponse{
			Speech: "Yes, I know we have " + wine.Name,
		}
	}

	// mark as available
	wine.Available = true
	err := env.db.Model(&wine).Where("id = ?", wine.Id).Updates(wine).Error
	if err != nil {
		return &apiai.ActionResponse{
			Speech: "Error", // TODO
		}
	}

	return &apiai.ActionResponse{
		Speech: "Noted, we now have " + wine.Name,
	}
}

func (env *Env) Intent_Wine_Query(req apiai.ActionRequest) *apiai.ActionResponse {
	wineDesc := req.Result.Parameters["wine-descriptor"].(string)
	wine := env.WineDescriptorLookup(wineDesc)

	if wine == nil {
		return WineNotFoundResponse()
	}

	if wine.Available {
		return &apiai.ActionResponse{
			Speech: "Yes, we do have " + wine.Name,
		}
	} else {
		return &apiai.ActionResponse{
			Speech: "No, we don't have any of the " + wine.Name + " available",
		}
	}
}

func (env *Env) ApiaiWebhookHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var req apiai.ActionRequest
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid request")
		return
	}

	intentHandlers := map[string](func(apiai.ActionRequest) *apiai.ActionResponse){
		"wine.list":               env.Intent_Wine_List,
		"wine.describe":           env.Intent_Wine_Describe,
		"wine.pair":               env.Intent_Wine_Pair,
		"wine.mark-unavailable":   env.Intent_Wine_Remove,
		"wine.mark-available":     env.Intent_Wine_Add,
		"wine.query-availability": env.Intent_Wine_Query,
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
