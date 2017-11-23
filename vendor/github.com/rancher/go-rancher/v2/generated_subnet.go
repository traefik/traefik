package client

const (
	SUBNET_TYPE = "subnet"
)

type Subnet struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	CidrSize int64 `json:"cidrSize,omitempty" yaml:"cidr_size,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	Data map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	EndAddress string `json:"endAddress,omitempty" yaml:"end_address,omitempty"`

	Gateway string `json:"gateway,omitempty" yaml:"gateway,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	NetworkAddress string `json:"networkAddress,omitempty" yaml:"network_address,omitempty"`

	NetworkId string `json:"networkId,omitempty" yaml:"network_id,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	StartAddress string `json:"startAddress,omitempty" yaml:"start_address,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	TransitioningProgress int64 `json:"transitioningProgress,omitempty" yaml:"transitioning_progress,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type SubnetCollection struct {
	Collection
	Data   []Subnet `json:"data,omitempty"`
	client *SubnetClient
}

type SubnetClient struct {
	rancherClient *RancherClient
}

type SubnetOperations interface {
	List(opts *ListOpts) (*SubnetCollection, error)
	Create(opts *Subnet) (*Subnet, error)
	Update(existing *Subnet, updates interface{}) (*Subnet, error)
	ById(id string) (*Subnet, error)
	Delete(container *Subnet) error

	ActionActivate(*Subnet) (*Subnet, error)

	ActionCreate(*Subnet) (*Subnet, error)

	ActionDeactivate(*Subnet) (*Subnet, error)

	ActionPurge(*Subnet) (*Subnet, error)

	ActionRemove(*Subnet) (*Subnet, error)

	ActionUpdate(*Subnet) (*Subnet, error)
}

func newSubnetClient(rancherClient *RancherClient) *SubnetClient {
	return &SubnetClient{
		rancherClient: rancherClient,
	}
}

func (c *SubnetClient) Create(container *Subnet) (*Subnet, error) {
	resp := &Subnet{}
	err := c.rancherClient.doCreate(SUBNET_TYPE, container, resp)
	return resp, err
}

func (c *SubnetClient) Update(existing *Subnet, updates interface{}) (*Subnet, error) {
	resp := &Subnet{}
	err := c.rancherClient.doUpdate(SUBNET_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *SubnetClient) List(opts *ListOpts) (*SubnetCollection, error) {
	resp := &SubnetCollection{}
	err := c.rancherClient.doList(SUBNET_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *SubnetCollection) Next() (*SubnetCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &SubnetCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *SubnetClient) ById(id string) (*Subnet, error) {
	resp := &Subnet{}
	err := c.rancherClient.doById(SUBNET_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *SubnetClient) Delete(container *Subnet) error {
	return c.rancherClient.doResourceDelete(SUBNET_TYPE, &container.Resource)
}

func (c *SubnetClient) ActionActivate(resource *Subnet) (*Subnet, error) {

	resp := &Subnet{}

	err := c.rancherClient.doAction(SUBNET_TYPE, "activate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *SubnetClient) ActionCreate(resource *Subnet) (*Subnet, error) {

	resp := &Subnet{}

	err := c.rancherClient.doAction(SUBNET_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *SubnetClient) ActionDeactivate(resource *Subnet) (*Subnet, error) {

	resp := &Subnet{}

	err := c.rancherClient.doAction(SUBNET_TYPE, "deactivate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *SubnetClient) ActionPurge(resource *Subnet) (*Subnet, error) {

	resp := &Subnet{}

	err := c.rancherClient.doAction(SUBNET_TYPE, "purge", &resource.Resource, nil, resp)

	return resp, err
}

func (c *SubnetClient) ActionRemove(resource *Subnet) (*Subnet, error) {

	resp := &Subnet{}

	err := c.rancherClient.doAction(SUBNET_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *SubnetClient) ActionUpdate(resource *Subnet) (*Subnet, error) {

	resp := &Subnet{}

	err := c.rancherClient.doAction(SUBNET_TYPE, "update", &resource.Resource, nil, resp)

	return resp, err
}
