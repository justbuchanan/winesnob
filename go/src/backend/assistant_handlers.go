package main

import (
    "fmt"
    "backend/apiai"
    "net/http"
    "encoding/json"
    "log"
)

// these values are set via the server config file
var (
    APIAI_AUTH_USERNAME = "apiai"
    APIAI_AUTH_PASSWORD = "?"
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
    fmt.Println(wine)
    if wine != nil {
        resp.Speech = wine.Name + ": " + wine.Description
    } else {
        resp.Speech = "I'm sorry, I couldn't find a wine matching that description"
    }

    return &resp
}

func Intent_Wine_Pair(req apiai.ActionRequest) *apiai.ActionResponse {
    food := req.Result.Parameters["food"].(string)
    return &apiai.ActionResponse{
        Speech: "I'd recommend the amarone, it goes very well with " + food,
    }
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
    // http basic auth
    user, pass, _ := r.BasicAuth()
    if !(user == APIAI_AUTH_USERNAME && pass == APIAI_AUTH_PASSWORD) {
        w.WriteHeader(http.StatusForbidden)
        return
    }

    decoder := json.NewDecoder(r.Body)
    var req apiai.ActionRequest
    err := decoder.Decode(&req)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprintf(w, "Invalid request")
        return
    }

    intentHandlers := map[string](func(apiai.ActionRequest) *apiai.ActionResponse) {
        "wine.list": Intent_Wine_List,
        "wine.describe": Intent_Wine_Describe,
        "wine.pair": Intent_Wine_Pair,
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
