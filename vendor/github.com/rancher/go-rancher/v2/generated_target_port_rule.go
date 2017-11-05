package client

const (
	TARGET_PORT_RULE_TYPE = "targetPortRule"
)

type TargetPortRule struct {
	Resource

	BackendName string `json:"backendName,omitempty" yaml:"backend_name,omitempty"`

	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`

	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	TargetPort int64 `json:"targetPort,omitempty" yaml:"target_port,omitempty"`
}

type TargetPortRuleCollection struct {
	Collection
	Data   []TargetPortRule `json:"data,omitempty"`
	client *TargetPortRuleClient
}

type TargetPortRuleClient struct {
	rancherClient *RancherClient
}

type TargetPortRuleOperations interface {
	List(opts *ListOpts) (*TargetPortRuleCollection, error)
	Create(opts *TargetPortRule) (*TargetPortRule, error)
	Update(existing *TargetPortRule, updates interface{}) (*TargetPortRule, error)
	ById(id string) (*TargetPortRule, error)
	Delete(container *TargetPortRule) error
}

func newTargetPortRuleClient(rancherClient *RancherClient) *TargetPortRuleClient {
	return &TargetPortRuleClient{
		rancherClient: rancherClient,
	}
}

func (c *TargetPortRuleClient) Create(container *TargetPortRule) (*TargetPortRule, error) {
	resp := &TargetPortRule{}
	err := c.rancherClient.doCreate(TARGET_PORT_RULE_TYPE, container, resp)
	return resp, err
}

func (c *TargetPortRuleClient) Update(existing *TargetPortRule, updates interface{}) (*TargetPortRule, error) {
	resp := &TargetPortRule{}
	err := c.rancherClient.doUpdate(TARGET_PORT_RULE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *TargetPortRuleClient) List(opts *ListOpts) (*TargetPortRuleCollection, error) {
	resp := &TargetPortRuleCollection{}
	err := c.rancherClient.doList(TARGET_PORT_RULE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *TargetPortRuleCollection) Next() (*TargetPortRuleCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &TargetPortRuleCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *TargetPortRuleClient) ById(id string) (*TargetPortRule, error) {
	resp := &TargetPortRule{}
	err := c.rancherClient.doById(TARGET_PORT_RULE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *TargetPortRuleClient) Delete(container *TargetPortRule) error {
	return c.rancherClient.doResourceDelete(TARGET_PORT_RULE_TYPE, &container.Resource)
}
