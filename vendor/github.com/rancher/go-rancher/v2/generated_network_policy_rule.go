package client

const (
	NETWORK_POLICY_RULE_TYPE = "networkPolicyRule"
)

type NetworkPolicyRule struct {
	Resource

	Action string `json:"action,omitempty" yaml:"action,omitempty"`

	Between NetworkPolicyRuleBetween `json:"between,omitempty" yaml:"between,omitempty"`

	From NetworkPolicyRuleMember `json:"from,omitempty" yaml:"from,omitempty"`

	Ports []string `json:"ports,omitempty" yaml:"ports,omitempty"`

	To NetworkPolicyRuleMember `json:"to,omitempty" yaml:"to,omitempty"`

	Within string `json:"within,omitempty" yaml:"within,omitempty"`
}

type NetworkPolicyRuleCollection struct {
	Collection
	Data   []NetworkPolicyRule `json:"data,omitempty"`
	client *NetworkPolicyRuleClient
}

type NetworkPolicyRuleClient struct {
	rancherClient *RancherClient
}

type NetworkPolicyRuleOperations interface {
	List(opts *ListOpts) (*NetworkPolicyRuleCollection, error)
	Create(opts *NetworkPolicyRule) (*NetworkPolicyRule, error)
	Update(existing *NetworkPolicyRule, updates interface{}) (*NetworkPolicyRule, error)
	ById(id string) (*NetworkPolicyRule, error)
	Delete(container *NetworkPolicyRule) error
}

func newNetworkPolicyRuleClient(rancherClient *RancherClient) *NetworkPolicyRuleClient {
	return &NetworkPolicyRuleClient{
		rancherClient: rancherClient,
	}
}

func (c *NetworkPolicyRuleClient) Create(container *NetworkPolicyRule) (*NetworkPolicyRule, error) {
	resp := &NetworkPolicyRule{}
	err := c.rancherClient.doCreate(NETWORK_POLICY_RULE_TYPE, container, resp)
	return resp, err
}

func (c *NetworkPolicyRuleClient) Update(existing *NetworkPolicyRule, updates interface{}) (*NetworkPolicyRule, error) {
	resp := &NetworkPolicyRule{}
	err := c.rancherClient.doUpdate(NETWORK_POLICY_RULE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *NetworkPolicyRuleClient) List(opts *ListOpts) (*NetworkPolicyRuleCollection, error) {
	resp := &NetworkPolicyRuleCollection{}
	err := c.rancherClient.doList(NETWORK_POLICY_RULE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *NetworkPolicyRuleCollection) Next() (*NetworkPolicyRuleCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &NetworkPolicyRuleCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *NetworkPolicyRuleClient) ById(id string) (*NetworkPolicyRule, error) {
	resp := &NetworkPolicyRule{}
	err := c.rancherClient.doById(NETWORK_POLICY_RULE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *NetworkPolicyRuleClient) Delete(container *NetworkPolicyRule) error {
	return c.rancherClient.doResourceDelete(NETWORK_POLICY_RULE_TYPE, &container.Resource)
}
