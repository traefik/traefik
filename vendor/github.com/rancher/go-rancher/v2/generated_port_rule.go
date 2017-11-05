package client

const (
	PORT_RULE_TYPE = "portRule"
)

type PortRule struct {
	Resource

	BackendName string `json:"backendName,omitempty" yaml:"backend_name,omitempty"`

	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`

	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	Priority int64 `json:"priority,omitempty" yaml:"priority,omitempty"`

	Protocol string `json:"protocol,omitempty" yaml:"protocol,omitempty"`

	Selector string `json:"selector,omitempty" yaml:"selector,omitempty"`

	ServiceId string `json:"serviceId,omitempty" yaml:"service_id,omitempty"`

	SourcePort int64 `json:"sourcePort,omitempty" yaml:"source_port,omitempty"`

	TargetPort int64 `json:"targetPort,omitempty" yaml:"target_port,omitempty"`
}

type PortRuleCollection struct {
	Collection
	Data   []PortRule `json:"data,omitempty"`
	client *PortRuleClient
}

type PortRuleClient struct {
	rancherClient *RancherClient
}

type PortRuleOperations interface {
	List(opts *ListOpts) (*PortRuleCollection, error)
	Create(opts *PortRule) (*PortRule, error)
	Update(existing *PortRule, updates interface{}) (*PortRule, error)
	ById(id string) (*PortRule, error)
	Delete(container *PortRule) error
}

func newPortRuleClient(rancherClient *RancherClient) *PortRuleClient {
	return &PortRuleClient{
		rancherClient: rancherClient,
	}
}

func (c *PortRuleClient) Create(container *PortRule) (*PortRule, error) {
	resp := &PortRule{}
	err := c.rancherClient.doCreate(PORT_RULE_TYPE, container, resp)
	return resp, err
}

func (c *PortRuleClient) Update(existing *PortRule, updates interface{}) (*PortRule, error) {
	resp := &PortRule{}
	err := c.rancherClient.doUpdate(PORT_RULE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *PortRuleClient) List(opts *ListOpts) (*PortRuleCollection, error) {
	resp := &PortRuleCollection{}
	err := c.rancherClient.doList(PORT_RULE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *PortRuleCollection) Next() (*PortRuleCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &PortRuleCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *PortRuleClient) ById(id string) (*PortRule, error) {
	resp := &PortRule{}
	err := c.rancherClient.doById(PORT_RULE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *PortRuleClient) Delete(container *PortRule) error {
	return c.rancherClient.doResourceDelete(PORT_RULE_TYPE, &container.Resource)
}
