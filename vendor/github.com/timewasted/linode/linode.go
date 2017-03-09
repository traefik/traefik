package linode

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	// ApiEndpoint is the base URL for the Linode API endpoint.
	ApiEndpoint = "https://api.linode.com/"
)

// The following are keys related to working with the Linode API.
const (
	apiAction       = "api_action"
	apiKey          = "api_key"
	apiRequestArray = "api_requestArray"
)

var (
	// client is an HTTP client which will be used for making API calls.
	client = &http.Client{
		Transport: &http.Transport{},
	}
)

type (
	// Action represents an action to be performed.
	Action struct {
		parameters Parameters
		result     interface{}
	}
	// Linode represents the interface to the Linode API.
	Linode struct {
		apiEndpoint string
		apiKey      string
		actions     []Action
	}
	// ResponseError represents an error returned by an API call.
	ResponseError struct {
		Code    int    `json:"ERRORCODE"`
		Message string `json:"ERRORMESSAGE"`
	}
	// Response represents the response to an API call.  Data is defined as
	// an interface, because each API call will return a different response.
	// It is the user's responsibility to turn it into something useful.
	Response struct {
		Action  string          `json:"ACTION"`
		RawData json.RawMessage `json:"DATA"`
		Errors  []ResponseError `json:"ERRORARRAY"`
		Data    interface{}     `json:"-"`
	}
)

// New returns a pointer to a new Linode object.
func New(apiKey string) *Linode {
	return &Linode{
		apiEndpoint: ApiEndpoint,
		apiKey:      apiKey,
		actions:     []Action{},
	}
}

// SetEndpoint sets the endpoint that all API requests will be sent to.  This
// should only be used for testing/debugging purposes!
func (l *Linode) SetEndpoint(endpoint string) {
	l.apiEndpoint = endpoint
}

// Request performs a single API operation and returns the full response.
func (l *Linode) Request(action string, params Parameters, result interface{}) (*Response, error) {
	// Initialize the request.
	req, err := http.NewRequest("GET", l.apiEndpoint, nil)
	if err != nil {
		return nil, NewError(err)
	}
	params.Set(apiAction, action)
	params.Set(apiKey, l.apiKey)
	req.URL.RawQuery = params.Encode()

	// Make the request.
	resp, err := client.Do(req)
	if err != nil {
		return nil, NewError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, NewError(fmt.Errorf("Expected status code %d, received %d", http.StatusOK, resp.StatusCode))
	}

	// Decode the response.
	var response *Response
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&response); err != nil {
		return nil, NewError(err)
	}

	// Decode the raw data into action-specific data.
	if result != nil {
		if err = json.Unmarshal(response.RawData, &result); err != nil {
			return nil, NewError(err)
		}
		response.Data = result
	}

	// If we have errors in the response, return both the response as well
	// as an API error.
	if len(response.Errors) != 0 {
		return response, NewApiError(response.Errors[0].Code, response.Errors[0].Message)
	}
	return response, nil
}

// Batch adds a new action to the batch request to be performed.
func (l *Linode) Batch(action string, params Parameters, result interface{}) {
	params.Set(apiAction, action)
	l.actions = append(l.actions, Action{
		parameters: params,
		result:     result,
	})
}

// Process performs all batch actions that have been added.
func (l *Linode) Process() ([]*Response, error) {
	// Quickly handle the case where we have zero or one actions to perform.
	count := len(l.actions)
	if count == 0 {
		return nil, NewError(errors.New("linode: request must contain at least one action"))
	}
	defer l.Reset()
	if count == 1 {
		resp, err := l.Request(l.actions[0].parameters.Get(apiAction), l.actions[0].parameters, l.actions[0].result)
		if resp == nil {
			return nil, err
		}
		return []*Response{resp}, err
	}

	// Prepare the parameters.
	params := make([]Parameters, 0, len(l.actions))
	for _, action := range l.actions {
		params = append(params, action.parameters)
	}
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return nil, NewError(err)
	}

	// Initialize the request.
	req, err := http.NewRequest("GET", l.apiEndpoint, nil)
	if err != nil {
		return nil, NewError(err)
	}
	query := Parameters{
		apiKey:          l.apiKey,
		apiAction:       "batch",
		apiRequestArray: string(jsonParams),
	}
	req.URL.RawQuery = query.Encode()

	// Do the request.
	resp, err := client.Do(req)
	if err != nil {
		return nil, NewError(err)
	}
	defer resp.Body.Close()

	// Decode the response.
	var response []*Response
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&response); err != nil {
		return nil, NewError(err)
	}

	// Decode the raw data into action-specific data.
	// FIXME: This whole block needs error handling, and I think since this
	// is handling multiple actions, it should probably handle errors on a
	// per-action basis.
	for index, action := range response {
		if l.actions[index].result == nil {
			continue
		}
		if err = json.Unmarshal(action.RawData, &l.actions[index].result); err != nil {
			continue
		}
		response[index].Data = l.actions[index].result
	}

	return response, nil
}

// Reset clears the list of actions to be performed.
func (l *Linode) Reset() {
	l.actions = []Action{}
}
