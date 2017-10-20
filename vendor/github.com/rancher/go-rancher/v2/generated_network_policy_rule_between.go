package client

const (
	NETWORK_POLICY_RULE_BETWEEN_TYPE = "networkPolicyRuleBetween"
)

type NetworkPolicyRuleBetween struct {
	Resource

	GroupBy string `json:"groupBy,omitempty" yaml:"group_by,omitempty"`

	Selector string `json:"selector,omitempty" yaml:"selector,omitempty"`
}

type NetworkPolicyRuleBetweenCollection struct {
	Collection
	Data   []NetworkPolicyRuleBetween `json:"data,omitempty"`
	client *NetworkPolicyRuleBetweenClient
}

type NetworkPolicyRuleBetweenClient struct {
	rancherClient *RancherClient
}

type NetworkPolicyRuleBetweenOperations interface {
	List(opts *ListOpts) (*NetworkPolicyRuleBetweenCollection, error)
	Create(opts *NetworkPolicyRuleBetween) (*NetworkPolicyRuleBetween, error)
	Update(existing *NetworkPolicyRuleBetween, updates interface{}) (*NetworkPolicyRuleBetween, error)
	ById(id string) (*NetworkPolicyRuleBetween, error)
	Delete(container *NetworkPolicyRuleBetween) error
}

func newNetworkPolicyRuleBetweenClient(rancherClient *RancherClient) *NetworkPolicyRuleBetweenClient {
	return &NetworkPolicyRuleBetweenClient{
		rancherClient: rancherClient,
	}
}

func (c *NetworkPolicyRuleBetweenClient) Create(container *NetworkPolicyRuleBetween) (*NetworkPolicyRuleBetween, error) {
	resp := &NetworkPolicyRuleBetween{}
	err := c.rancherClient.doCreate(NETWORK_POLICY_RULE_BETWEEN_TYPE, container, resp)
	return resp, err
}

func (c *NetworkPolicyRuleBetweenClient) Update(existing *NetworkPolicyRuleBetween, updates interface{}) (*NetworkPolicyRuleBetween, error) {
	resp := &NetworkPolicyRuleBetween{}
	err := c.rancherClient.doUpdate(NETWORK_POLICY_RULE_BETWEEN_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *NetworkPolicyRuleBetweenClient) List(opts *ListOpts) (*NetworkPolicyRuleBetweenCollection, error) {
	resp := &NetworkPolicyRuleBetweenCollection{}
	err := c.rancherClient.doList(NETWORK_POLICY_RULE_BETWEEN_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *NetworkPolicyRuleBetweenCollection) Next() (*NetworkPolicyRuleBetweenCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &NetworkPolicyRuleBetweenCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *NetworkPolicyRuleBetweenClient) ById(id string) (*NetworkPolicyRuleBetween, error) {
	resp := &NetworkPolicyRuleBetween{}
	err := c.rancherClient.doById(NETWORK_POLICY_RULE_BETWEEN_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *NetworkPolicyRuleBetweenClient) Delete(container *NetworkPolicyRuleBetween) error {
	return c.rancherClient.doResourceDelete(NETWORK_POLICY_RULE_BETWEEN_TYPE, &container.Resource)
}
