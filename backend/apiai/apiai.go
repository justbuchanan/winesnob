package apiai

import (
	"time"
)

type ActionResponse struct {
	Speech      string `json:"speech"`
	DisplayText string `json:"displayText"`
	Data        struct {
	} `json:"data"`
	ContextOut []string `json:"contextOut"`
	Source     string   `json:"source"`
}

type ActionRequest struct {
	ID              string      `json:"id"`
	OriginalRequest interface{} `json:"originalRequest"`
	Result          struct {
		Action           string        `json:"action"`
		ActionIncomplete bool          `json:"actionIncomplete"`
		Contexts         []interface{} `json:"contexts"`
		Fulfillment      struct {
			Messages []struct {
				Speech string `json:"speech"`
				Type   int    `json:"type"`
			} `json:"messages"`
			Speech string `json:"speech"`
		} `json:"fulfillment"`
		Metadata struct {
			IntentID                  string `json:"intentId"`
			IntentName                string `json:"intentName"`
			WebhookForSlotFillingUsed string `json:"webhookForSlotFillingUsed"`
			WebhookUsed               string `json:"webhookUsed"`
		} `json:"metadata"`
		Parameters    map[string]interface{} `json:"parameters"`
		ResolvedQuery string                 `json:"resolvedQuery"`
		Score         float64                `json:"score"`
		Source        string                 `json:"source"`
		Speech        string                 `json:"speech"`
	} `json:"result"`
	SessionID string `json:"sessionId"`
	Status    struct {
		Code      int    `json:"code"`
		ErrorType string `json:"errorType"`
	} `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}
