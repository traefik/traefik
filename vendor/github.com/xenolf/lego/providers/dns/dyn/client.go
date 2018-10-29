package dyn

import "encoding/json"

const defaultBaseURL = "https://api.dynect.net/REST"

type dynResponse struct {
	// One of 'success', 'failure', or 'incomplete'
	Status string `json:"status"`

	// The structure containing the actual results of the request
	Data json.RawMessage `json:"data"`

	// The ID of the job that was created in response to a request.
	JobID int `json:"job_id"`

	// A list of zero or more messages
	Messages json.RawMessage `json:"msgs"`
}

type creds struct {
	Customer string `json:"customer_name"`
	User     string `json:"user_name"`
	Pass     string `json:"password"`
}

type session struct {
	Token   string `json:"token"`
	Version string `json:"version"`
}

type publish struct {
	Publish bool   `json:"publish"`
	Notes   string `json:"notes"`
}
