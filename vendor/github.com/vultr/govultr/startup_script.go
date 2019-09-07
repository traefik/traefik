package govultr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// StartupScriptService is the interface to interact with the startup script endpoints on the Vultr API
// Link: https://www.vultr.com/api/#startupscript
type StartupScriptService interface {
	Create(ctx context.Context, name, script, scriptType string) (*StartupScript, error)
	Delete(ctx context.Context, scriptID string) error
	List(ctx context.Context) ([]StartupScript, error)
	Update(ctx context.Context, script *StartupScript) error
}

// StartupScriptServiceHandler handles interaction with the startup script methods for the Vultr API
type StartupScriptServiceHandler struct {
	client *Client
}

// StartupScript represents an startup script on Vultr
type StartupScript struct {
	ScriptID     string `json:"SCRIPTID"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Script       string `json:"script"`
}

// UnmarshalJSON implements json.Unmarshaller on StartupScript to handle the inconsistent types returned from the Vultr API.
func (s *StartupScript) UnmarshalJSON(data []byte) (err error) {
	if s == nil {
		*s = StartupScript{}
	}

	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	s.ScriptID = fmt.Sprintf("%v", v["SCRIPTID"])
	s.DateCreated = fmt.Sprintf("%v", v["date_created"])
	s.DateModified = fmt.Sprintf("%v", v["date_modified"])
	s.Name = fmt.Sprintf("%v", v["name"])
	s.Type = fmt.Sprintf("%v", v["type"])
	s.Script = fmt.Sprintf("%v", v["script"])

	return nil
}

// Create will add the specified startup script to your Vultr account
func (s *StartupScriptServiceHandler) Create(ctx context.Context, name, script, scriptType string) (*StartupScript, error) {

	uri := "/v1/startupscript/create"

	values := url.Values{
		"name":   {name},
		"script": {script},
	}

	if scriptType != "" {
		values.Add("type", scriptType)
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return nil, err
	}

	ss := new(StartupScript)

	err = s.client.DoWithContext(ctx, req, ss)

	if err != nil {
		return nil, err
	}

	ss.DateCreated = ""
	ss.DateModified = ""
	ss.Name = name
	ss.Type = scriptType
	ss.Script = script

	return ss, nil
}

// Delete will delete the specified startup script from your Vultr account
func (s *StartupScriptServiceHandler) Delete(ctx context.Context, scriptID string) error {

	uri := "/v1/startupscript/destroy"

	values := url.Values{
		"SCRIPTID": {scriptID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// List will list all the startup scripts associated with your Vultr account
func (s *StartupScriptServiceHandler) List(ctx context.Context) ([]StartupScript, error) {

	uri := "/v1/startupscript/list"

	req, err := s.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	scriptsMap := make(map[string]StartupScript)
	err = s.client.DoWithContext(ctx, req, &scriptsMap)
	if err != nil {
		return nil, err
	}

	var scripts []StartupScript
	for _, key := range scriptsMap {
		scripts = append(scripts, key)
	}

	return scripts, nil
}

// Update will update the given startup script. Empty strings will be ignored.
func (s *StartupScriptServiceHandler) Update(ctx context.Context, script *StartupScript) error {

	uri := "/v1/startupscript/update"

	values := url.Values{
		"SCRIPTID": {script.ScriptID},
	}

	// Optional
	if script.Name != "" {
		values.Add("name", script.Name)
	}
	if script.Script != "" {
		values.Add("script", script.Script)
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}
