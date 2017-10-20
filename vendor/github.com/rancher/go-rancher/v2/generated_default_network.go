package client

const (
	DEFAULT_NETWORK_TYPE = "defaultNetwork"
)

type DefaultNetwork struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	Data map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`

	DefaultPolicyAction string `json:"defaultPolicyAction,omitempty" yaml:"default_policy_action,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	Dns []string `json:"dns,omitempty" yaml:"dns,omitempty"`

	DnsSearch []string `json:"dnsSearch,omitempty" yaml:"dns_search,omitempty"`

	HostPorts bool `json:"hostPorts,omitempty" yaml:"host_ports,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	Policy []NetworkPolicyRule `json:"policy,omitempty" yaml:"policy,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Subnets []Subnet `json:"subnets,omitempty" yaml:"subnets,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	TransitioningProgress int64 `json:"transitioningProgress,omitempty" yaml:"transitioning_progress,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type DefaultNetworkCollection struct {
	Collection
	Data   []DefaultNetwork `json:"data,omitempty"`
	client *DefaultNetworkClient
}

type DefaultNetworkClient struct {
	rancherClient *RancherClient
}

type DefaultNetworkOperations interface {
	List(opts *ListOpts) (*DefaultNetworkCollection, error)
	Create(opts *DefaultNetwork) (*DefaultNetwork, error)
	Update(existing *DefaultNetwork, updates interface{}) (*DefaultNetwork, error)
	ById(id string) (*DefaultNetwork, error)
	Delete(container *DefaultNetwork) error

	ActionActivate(*DefaultNetwork) (*Network, error)

	ActionCreate(*DefaultNetwork) (*Network, error)

	ActionDeactivate(*DefaultNetwork) (*Network, error)

	ActionPurge(*DefaultNetwork) (*Network, error)

	ActionRemove(*DefaultNetwork) (*Network, error)

	ActionUpdate(*DefaultNetwork) (*Network, error)
}

func newDefaultNetworkClient(rancherClient *RancherClient) *DefaultNetworkClient {
	return &DefaultNetworkClient{
		rancherClient: rancherClient,
	}
}

func (c *DefaultNetworkClient) Create(container *DefaultNetwork) (*DefaultNetwork, error) {
	resp := &DefaultNetwork{}
	err := c.rancherClient.doCreate(DEFAULT_NETWORK_TYPE, container, resp)
	return resp, err
}

func (c *DefaultNetworkClient) Update(existing *DefaultNetwork, updates interface{}) (*DefaultNetwork, error) {
	resp := &DefaultNetwork{}
	err := c.rancherClient.doUpdate(DEFAULT_NETWORK_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *DefaultNetworkClient) List(opts *ListOpts) (*DefaultNetworkCollection, error) {
	resp := &DefaultNetworkCollection{}
	err := c.rancherClient.doList(DEFAULT_NETWORK_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *DefaultNetworkCollection) Next() (*DefaultNetworkCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &DefaultNetworkCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *DefaultNetworkClient) ById(id string) (*DefaultNetwork, error) {
	resp := &DefaultNetwork{}
	err := c.rancherClient.doById(DEFAULT_NETWORK_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *DefaultNetworkClient) Delete(container *DefaultNetwork) error {
	return c.rancherClient.doResourceDelete(DEFAULT_NETWORK_TYPE, &container.Resource)
}

func (c *DefaultNetworkClient) ActionActivate(resource *DefaultNetwork) (*Network, error) {

	resp := &Network{}

	err := c.rancherClient.doAction(DEFAULT_NETWORK_TYPE, "activate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DefaultNetworkClient) ActionCreate(resource *DefaultNetwork) (*Network, error) {

	resp := &Network{}

	err := c.rancherClient.doAction(DEFAULT_NETWORK_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DefaultNetworkClient) ActionDeactivate(resource *DefaultNetwork) (*Network, error) {

	resp := &Network{}

	err := c.rancherClient.doAction(DEFAULT_NETWORK_TYPE, "deactivate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DefaultNetworkClient) ActionPurge(resource *DefaultNetwork) (*Network, error) {

	resp := &Network{}

	err := c.rancherClient.doAction(DEFAULT_NETWORK_TYPE, "purge", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DefaultNetworkClient) ActionRemove(resource *DefaultNetwork) (*Network, error) {

	resp := &Network{}

	err := c.rancherClient.doAction(DEFAULT_NETWORK_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DefaultNetworkClient) ActionUpdate(resource *DefaultNetwork) (*Network, error) {

	resp := &Network{}

	err := c.rancherClient.doAction(DEFAULT_NETWORK_TYPE, "update", &resource.Resource, nil, resp)

	return resp, err
}
