package namecom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
)

var _ = bytes.MinRead

// HelloFunc returns some information about the API server.
func (n *NameCom) HelloFunc(request *HelloRequest) (*HelloResponse, error) {
	endpoint := fmt.Sprintf("/v4/hello")

	values := url.Values{}

	body, err := n.get(endpoint, values)
	if err != nil {
		return nil, err
	}

	resp := &HelloResponse{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
