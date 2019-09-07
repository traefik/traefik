package govultr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// OSService is the interface to interact with the operating system endpoint on the Vultr API
// Link: https://www.vultr.com/api/#os
type OSService interface {
	List(ctx context.Context) ([]OS, error)
}

// OSServiceHandler handles interaction with the operating system methods for the Vultr API
type OSServiceHandler struct {
	client *Client
}

// OS represents a Vultr operating system
type OS struct {
	OsID    int    `json:"OSID"`
	Name    string `json:"name"`
	Arch    string `json:"arch"`
	Family  string `json:"family"`
	Windows bool   `json:"windows"`
}

// UnmarshalJSON implements json.Unmarshaller on OS to handle the inconsistent types returned from the Vultr API.
func (o *OS) UnmarshalJSON(data []byte) (err error) {
	if o == nil {
		*o = OS{}
	}

	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	i, err := strconv.Atoi(fmt.Sprintf("%v", v["OSID"]))
	if err != nil {
		return err
	}
	o.OsID = i

	value := fmt.Sprintf("%v", v["windows"])
	o.Windows = false
	if value == "true" {
		o.Windows = true
	}

	o.Name = fmt.Sprintf("%v", v["name"])
	o.Arch = fmt.Sprintf("%v", v["arch"])
	o.Family = fmt.Sprintf("%v", v["family"])

	return nil
}

// List retrieves a list of available operating systems.
// If the Windows flag is true, a Windows license will be included with the instance, which will increase the cost.
func (o *OSServiceHandler) List(ctx context.Context) ([]OS, error) {
	uri := "/v1/os/list"
	req, err := o.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	osMap := make(map[string]OS)

	err = o.client.DoWithContext(ctx, req, &osMap)
	if err != nil {
		return nil, err
	}

	var oses []OS
	for _, os := range osMap {
		oses = append(oses, os)
	}

	return oses, nil
}
