package lib

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// StartupScript on Vultr account
type StartupScript struct {
	ID      string `json:"SCRIPTID"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"script"`
}

type startupscripts []StartupScript

func (s startupscripts) Len() int      { return len(s) }
func (s startupscripts) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s startupscripts) Less(i, j int) bool {
	return strings.ToLower(s[i].Name) < strings.ToLower(s[j].Name)
}

// UnmarshalJSON implements json.Unmarshaller on StartupScript.
// Necessary because the SCRIPTID field has inconsistent types.
func (s *StartupScript) UnmarshalJSON(data []byte) (err error) {
	if s == nil {
		*s = StartupScript{}
	}

	var fields map[string]interface{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	s.ID = fmt.Sprintf("%v", fields["SCRIPTID"])
	s.Name = fmt.Sprintf("%v", fields["name"])
	s.Type = fmt.Sprintf("%v", fields["type"])
	s.Content = fmt.Sprintf("%v", fields["script"])

	return
}

// GetStartupScripts returns a list of all startup scripts on the current Vultr account
func (c *Client) GetStartupScripts() (scripts []StartupScript, err error) {
	var scriptMap map[string]StartupScript
	if err := c.get(`startupscript/list`, &scriptMap); err != nil {
		return nil, err
	}

	for _, script := range scriptMap {
		if script.Type == "" {
			script.Type = "boot" // set default script type
		}
		scripts = append(scripts, script)
	}
	sort.Sort(startupscripts(scripts))
	return scripts, nil
}

// GetStartupScript returns the startup script with the given ID
func (c *Client) GetStartupScript(id string) (StartupScript, error) {
	scripts, err := c.GetStartupScripts()
	if err != nil {
		return StartupScript{}, err
	}

	for _, s := range scripts {
		if s.ID == id {
			return s, nil
		}
	}
	return StartupScript{}, nil
}

// CreateStartupScript creates a new startup script
func (c *Client) CreateStartupScript(name, content, scriptType string) (StartupScript, error) {
	values := url.Values{
		"name":   {name},
		"script": {content},
		"type":   {scriptType},
	}

	var script StartupScript
	if err := c.post(`startupscript/create`, values, &script); err != nil {
		return StartupScript{}, err
	}
	script.Name = name
	script.Content = content
	script.Type = scriptType

	return script, nil
}

// UpdateStartupScript updates an existing startup script
func (c *Client) UpdateStartupScript(script StartupScript) error {
	values := url.Values{
		"SCRIPTID": {script.ID},
	}
	if script.Name != "" {
		values.Add("name", script.Name)
	}
	if script.Content != "" {
		values.Add("script", script.Content)
	}

	if err := c.post(`startupscript/update`, values, nil); err != nil {
		return err
	}
	return nil
}

// DeleteStartupScript deletes an existing startup script from Vultr account
func (c *Client) DeleteStartupScript(id string) error {
	values := url.Values{
		"SCRIPTID": {id},
	}

	if err := c.post(`startupscript/destroy`, values, nil); err != nil {
		return err
	}
	return nil
}
