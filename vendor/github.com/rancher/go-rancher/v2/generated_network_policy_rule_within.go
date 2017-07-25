package client

const (
	NETWORK_POLICY_RULE_WITHIN_TYPE = "networkPolicyRuleWithin"
)

type NetworkPolicyRuleWithin struct {
	Resource
}

type NetworkPolicyRuleWithinCollection struct {
	Collection
	Data   []NetworkPolicyRuleWithin `json:"data,omitempty"`
	client *NetworkPolicyRuleWithinClient
}

type NetworkPolicyRuleWithinClient struct {
	rancherClient *RancherClient
}

type NetworkPolicyRuleWithinOperations interface {
	List(opts *ListOpts) (*NetworkPolicyRuleWithinCollection, error)
	Create(opts *NetworkPolicyRuleWithin) (*NetworkPolicyRuleWithin, error)
	Update(existing *NetworkPolicyRuleWithin, updates interface{}) (*NetworkPolicyRuleWithin, error)
	ById(id string) (*NetworkPolicyRuleWithin, error)
	Delete(container *NetworkPolicyRuleWithin) error
}

func newNetworkPolicyRuleWithinClient(rancherClient *RancherClient) *NetworkPolicyRuleWithinClient {
	return &NetworkPolicyRuleWithinClient{
		rancherClient: rancherClient,
	}
}

func (c *NetworkPolicyRuleWithinClient) Create(container *NetworkPolicyRuleWithin) (*NetworkPolicyRuleWithin, error) {
	resp := &NetworkPolicyRuleWithin{}
	err := c.rancherClient.doCreate(NETWORK_POLICY_RULE_WITHIN_TYPE, container, resp)
	return resp, err
}

func (c *NetworkPolicyRuleWithinClient) Update(existing *NetworkPolicyRuleWithin, updates interface{}) (*NetworkPolicyRuleWithin, error) {
	resp := &NetworkPolicyRuleWithin{}
	err := c.rancherClient.doUpdate(NETWORK_POLICY_RULE_WITHIN_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *NetworkPolicyRuleWithinClient) List(opts *ListOpts) (*NetworkPolicyRuleWithinCollection, error) {
	resp := &NetworkPolicyRuleWithinCollection{}
	err := c.rancherClient.doList(NETWORK_POLICY_RULE_WITHIN_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *NetworkPolicyRuleWithinCollection) Next() (*NetworkPolicyRuleWithinCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &NetworkPolicyRuleWithinCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *NetworkPolicyRuleWithinClient) ById(id string) (*NetworkPolicyRuleWithin, error) {
	resp := &NetworkPolicyRuleWithin{}
	err := c.rancherClient.doById(NETWORK_POLICY_RULE_WITHIN_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *NetworkPolicyRuleWithinClient) Delete(container *NetworkPolicyRuleWithin) error {
	return c.rancherClient.doResourceDelete(NETWORK_POLICY_RULE_WITHIN_TYPE, &container.Resource)
}
