package main

import (
    "fmt"
    "backend/apiai"
    "net/http"
    "encoding/json"
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
    var resp apiai.ActionResponse

    wineDesc := req.Result.Parameters["wine-descriptor"].(string)
    wine := WineDescriptorLookup(wineDesc)
    if wine != nil {
        resp.Speech = wine.Name + ": " + wine.Description
    } else {
        resp.Speech = "I'm sorry, I couldn't find a wine matching that description"
    }

    return &resp
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
    decoder := json.NewDecoder(r.Body)
    var req apiai.ActionRequest
    err := decoder.Decode(&req)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprintf(w, "Invalid request")
        return
    }

    var resp *apiai.ActionResponse

    intent := req.Result.Metadata.IntentName
    if intent == "wine.list" {
        resp = Intent_Wine_List(req)
    } else if intent == "wine.describe" {
        resp = Intent_Wine_Describe(req)
    } else if intent == "wine.pair" {
        food := req.Result.Parameters["food"].(string)
        resp.Speech = "I'd recommend the amarone, it goes very well with " + food
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}
