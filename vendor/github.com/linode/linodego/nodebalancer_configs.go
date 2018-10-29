package linodego

import (
	"context"
	"encoding/json"
	"fmt"
)

// NodeBalancerConfig objects allow a NodeBalancer to accept traffic on a new port
type NodeBalancerConfig struct {
	ID             int                     `json:"id"`
	Port           int                     `json:"port"`
	Protocol       ConfigProtocol          `json:"protocol"`
	Algorithm      ConfigAlgorithm         `json:"algorithm"`
	Stickiness     ConfigStickiness        `json:"stickiness"`
	Check          ConfigCheck             `json:"check"`
	CheckInterval  int                     `json:"check_interval"`
	CheckAttempts  int                     `json:"check_attempts"`
	CheckPath      string                  `json:"check_path"`
	CheckBody      string                  `json:"check_body"`
	CheckPassive   bool                    `json:"check_passive"`
	CheckTimeout   int                     `json:"check_timeout"`
	CipherSuite    ConfigCipher            `json:"cipher_suite"`
	NodeBalancerID int                     `json:"nodebalancer_id"`
	SSLCommonName  string                  `json:"ssl_commonname"`
	SSLFingerprint string                  `json:"ssl_fingerprint"`
	SSLCert        string                  `json:"ssl_cert"`
	SSLKey         string                  `json:"ssl_key"`
	NodesStatus    *NodeBalancerNodeStatus `json:"nodes_status"`
}

// ConfigAlgorithm constants start with Algorithm and include Linode API NodeBalancer Config Algorithms
type ConfigAlgorithm string

// ConfigAlgorithm constants reflect the NodeBalancer Config Algorithm
const (
	AlgorithmRoundRobin ConfigAlgorithm = "roundrobin"
	AlgorithmLeastConn  ConfigAlgorithm = "leastconn"
	AlgorithmSource     ConfigAlgorithm = "source"
)

// ConfigStickiness constants start with Stickiness and include Linode API NodeBalancer Config Stickiness
type ConfigStickiness string

// ConfigStickiness constants reflect the node stickiness method for a NodeBalancer Config
const (
	StickinessNone       ConfigStickiness = "none"
	StickinessTable      ConfigStickiness = "table"
	StickinessHTTPCookie ConfigStickiness = "http_cookie"
)

// ConfigCheck constants start with Check and include Linode API NodeBalancer Config Check methods
type ConfigCheck string

// ConfigCheck constants reflect the node health status checking method for a NodeBalancer Config
const (
	CheckNone       ConfigCheck = "none"
	CheckConnection ConfigCheck = "connection"
	CheckHTTP       ConfigCheck = "http"
	CheckHTTPBody   ConfigCheck = "http_body"
)

// ConfigProtocol constants start with Protocol and include Linode API Nodebalancer Config protocols
type ConfigProtocol string

// ConfigProtocol constants reflect the protocol used by a NodeBalancer Config
const (
	ProtocolHTTP  ConfigProtocol = "http"
	ProtocolHTTPS ConfigProtocol = "https"
	ProtocolTCP   ConfigProtocol = "tcp"
)

// ConfigCipher constants start with Cipher and include Linode API NodeBalancer Config Cipher values
type ConfigCipher string

// ConfigCipher constants reflect the preferred cipher set for a NodeBalancer Config
const (
	CipherRecommended ConfigCipher = "recommended"
	CipherLegacy      ConfigCipher = "legacy"
)

// NodeBalancerNodeStatus represents the total number of nodes whose status is Up or Down
type NodeBalancerNodeStatus struct {
	Up   int `json:"up"`
	Down int `json:"down"`
}

// NodeBalancerConfigCreateOptions are permitted by CreateNodeBalancerConfig
type NodeBalancerConfigCreateOptions struct {
	Port          int                             `json:"port"`
	Protocol      ConfigProtocol                  `json:"protocol,omitempty"`
	Algorithm     ConfigAlgorithm                 `json:"algorithm,omitempty"`
	Stickiness    ConfigStickiness                `json:"stickiness,omitempty"`
	Check         ConfigCheck                     `json:"check,omitempty"`
	CheckInterval int                             `json:"check_interval,omitempty"`
	CheckAttempts int                             `json:"check_attempts,omitempty"`
	CheckPath     string                          `json:"check_path,omitempty"`
	CheckBody     string                          `json:"check_body,omitempty"`
	CheckPassive  *bool                           `json:"check_passive,omitempty"`
	CheckTimeout  int                             `json:"check_timeout,omitempty"`
	CipherSuite   ConfigCipher                    `json:"cipher_suite,omitempty"`
	SSLCert       string                          `json:"ssl_cert,omitempty"`
	SSLKey        string                          `json:"ssl_key,omitempty"`
	Nodes         []NodeBalancerNodeCreateOptions `json:"nodes,omitempty"`
}

// NodeBalancerConfigRebuildOptions used by RebuildNodeBalancerConfig
type NodeBalancerConfigRebuildOptions struct {
	Port          int                             `json:"port"`
	Protocol      ConfigProtocol                  `json:"protocol,omitempty"`
	Algorithm     ConfigAlgorithm                 `json:"algorithm,omitempty"`
	Stickiness    ConfigStickiness                `json:"stickiness,omitempty"`
	Check         ConfigCheck                     `json:"check,omitempty"`
	CheckInterval int                             `json:"check_interval,omitempty"`
	CheckAttempts int                             `json:"check_attempts,omitempty"`
	CheckPath     string                          `json:"check_path,omitempty"`
	CheckBody     string                          `json:"check_body,omitempty"`
	CheckPassive  *bool                           `json:"check_passive,omitempty"`
	CheckTimeout  int                             `json:"check_timeout,omitempty"`
	CipherSuite   ConfigCipher                    `json:"cipher_suite,omitempty"`
	SSLCert       string                          `json:"ssl_cert,omitempty"`
	SSLKey        string                          `json:"ssl_key,omitempty"`
	Nodes         []NodeBalancerNodeCreateOptions `json:"nodes"`
}

// NodeBalancerConfigUpdateOptions are permitted by UpdateNodeBalancerConfig
type NodeBalancerConfigUpdateOptions NodeBalancerConfigCreateOptions

// GetCreateOptions converts a NodeBalancerConfig to NodeBalancerConfigCreateOptions for use in CreateNodeBalancerConfig
func (i NodeBalancerConfig) GetCreateOptions() NodeBalancerConfigCreateOptions {
	return NodeBalancerConfigCreateOptions{
		Port:          i.Port,
		Protocol:      i.Protocol,
		Algorithm:     i.Algorithm,
		Stickiness:    i.Stickiness,
		Check:         i.Check,
		CheckInterval: i.CheckInterval,
		CheckAttempts: i.CheckAttempts,
		CheckTimeout:  i.CheckTimeout,
		CheckPath:     i.CheckPath,
		CheckBody:     i.CheckBody,
		CheckPassive:  &i.CheckPassive,
		CipherSuite:   i.CipherSuite,
		SSLCert:       i.SSLCert,
		SSLKey:        i.SSLKey,
	}
}

// GetUpdateOptions converts a NodeBalancerConfig to NodeBalancerConfigUpdateOptions for use in UpdateNodeBalancerConfig
func (i NodeBalancerConfig) GetUpdateOptions() NodeBalancerConfigUpdateOptions {
	return NodeBalancerConfigUpdateOptions{
		Port:          i.Port,
		Protocol:      i.Protocol,
		Algorithm:     i.Algorithm,
		Stickiness:    i.Stickiness,
		Check:         i.Check,
		CheckInterval: i.CheckInterval,
		CheckAttempts: i.CheckAttempts,
		CheckPath:     i.CheckPath,
		CheckBody:     i.CheckBody,
		CheckPassive:  &i.CheckPassive,
		CheckTimeout:  i.CheckTimeout,
		CipherSuite:   i.CipherSuite,
		SSLCert:       i.SSLCert,
		SSLKey:        i.SSLKey,
	}
}

// GetRebuildOptions converts a NodeBalancerConfig to NodeBalancerConfigRebuildOptions for use in RebuildNodeBalancerConfig
func (i NodeBalancerConfig) GetRebuildOptions() NodeBalancerConfigRebuildOptions {
	return NodeBalancerConfigRebuildOptions{
		Port:          i.Port,
		Protocol:      i.Protocol,
		Algorithm:     i.Algorithm,
		Stickiness:    i.Stickiness,
		Check:         i.Check,
		CheckInterval: i.CheckInterval,
		CheckAttempts: i.CheckAttempts,
		CheckTimeout:  i.CheckTimeout,
		CheckPath:     i.CheckPath,
		CheckBody:     i.CheckBody,
		CheckPassive:  &i.CheckPassive,
		CipherSuite:   i.CipherSuite,
		SSLCert:       i.SSLCert,
		SSLKey:        i.SSLKey,
		Nodes:         make([]NodeBalancerNodeCreateOptions, 0),
	}
}

// NodeBalancerConfigsPagedResponse represents a paginated NodeBalancerConfig API response
type NodeBalancerConfigsPagedResponse struct {
	*PageOptions
	Data []NodeBalancerConfig `json:"data"`
}

// endpointWithID gets the endpoint URL for NodeBalancerConfig
func (NodeBalancerConfigsPagedResponse) endpointWithID(c *Client, id int) string {
	endpoint, err := c.NodeBalancerConfigs.endpointWithID(id)
	if err != nil {
		panic(err)
	}
	return endpoint
}

// appendData appends NodeBalancerConfigs when processing paginated NodeBalancerConfig responses
func (resp *NodeBalancerConfigsPagedResponse) appendData(r *NodeBalancerConfigsPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListNodeBalancerConfigs lists NodeBalancerConfigs
func (c *Client) ListNodeBalancerConfigs(ctx context.Context, nodebalancerID int, opts *ListOptions) ([]NodeBalancerConfig, error) {
	response := NodeBalancerConfigsPagedResponse{}
	err := c.listHelperWithID(ctx, &response, nodebalancerID, opts)
	for i := range response.Data {
		response.Data[i].fixDates()
	}
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

// fixDates converts JSON timestamps to Go time.Time values
func (i *NodeBalancerConfig) fixDates() *NodeBalancerConfig {
	return i
}

// GetNodeBalancerConfig gets the template with the provided ID
func (c *Client) GetNodeBalancerConfig(ctx context.Context, nodebalancerID int, configID int) (*NodeBalancerConfig, error) {
	e, err := c.NodeBalancerConfigs.endpointWithID(nodebalancerID)
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, configID)
	r, err := coupleAPIErrors(c.R(ctx).SetResult(&NodeBalancerConfig{}).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*NodeBalancerConfig).fixDates(), nil
}

// CreateNodeBalancerConfig creates a NodeBalancerConfig
func (c *Client) CreateNodeBalancerConfig(ctx context.Context, nodebalancerID int, nodebalancerConfig NodeBalancerConfigCreateOptions) (*NodeBalancerConfig, error) {
	var body string
	e, err := c.NodeBalancerConfigs.endpointWithID(nodebalancerID)

	if err != nil {
		return nil, err
	}

	req := c.R(ctx).SetResult(&NodeBalancerConfig{})

	if bodyData, err := json.Marshal(nodebalancerConfig); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}
	return r.Result().(*NodeBalancerConfig).fixDates(), nil
}

// UpdateNodeBalancerConfig updates the NodeBalancerConfig with the specified id
func (c *Client) UpdateNodeBalancerConfig(ctx context.Context, nodebalancerID int, configID int, updateOpts NodeBalancerConfigUpdateOptions) (*NodeBalancerConfig, error) {
	var body string
	e, err := c.NodeBalancerConfigs.endpointWithID(nodebalancerID)
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, configID)

	req := c.R(ctx).SetResult(&NodeBalancerConfig{})

	if bodyData, err := json.Marshal(updateOpts); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Put(e))

	if err != nil {
		return nil, err
	}
	return r.Result().(*NodeBalancerConfig).fixDates(), nil
}

// DeleteNodeBalancerConfig deletes the NodeBalancerConfig with the specified id
func (c *Client) DeleteNodeBalancerConfig(ctx context.Context, nodebalancerID int, configID int) error {
	e, err := c.NodeBalancerConfigs.endpointWithID(nodebalancerID)
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%d", e, configID)

	_, err = coupleAPIErrors(c.R(ctx).Delete(e))
	return err
}

// RebuildNodeBalancerConfig updates the NodeBalancer with the specified id
func (c *Client) RebuildNodeBalancerConfig(ctx context.Context, nodeBalancerID int, configID int, rebuildOpts NodeBalancerConfigRebuildOptions) (*NodeBalancerConfig, error) {
	var body string
	e, err := c.NodeBalancerConfigs.endpointWithID(nodeBalancerID)
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d/rebuild", e, configID)

	req := c.R(ctx).SetResult(&NodeBalancerConfig{})

	if bodyData, err := json.Marshal(rebuildOpts); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}
	return r.Result().(*NodeBalancerConfig).fixDates(), nil
}
