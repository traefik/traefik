package client

const (
	STACK_UPGRADE_TYPE = "stackUpgrade"
)

type StackUpgrade struct {
	Resource

	Answers map[string]interface{} `json:"answers,omitempty" yaml:"answers,omitempty"`

	DockerCompose string `json:"dockerCompose,omitempty" yaml:"docker_compose,omitempty"`

	Environment map[string]interface{} `json:"environment,omitempty" yaml:"environment,omitempty"`

	ExternalId string `json:"externalId,omitempty" yaml:"external_id,omitempty"`

	RancherCompose string `json:"rancherCompose,omitempty" yaml:"rancher_compose,omitempty"`

	Templates map[string]interface{} `json:"templates,omitempty" yaml:"templates,omitempty"`
}

type StackUpgradeCollection struct {
	Collection
	Data   []StackUpgrade `json:"data,omitempty"`
	client *StackUpgradeClient
}

type StackUpgradeClient struct {
	rancherClient *RancherClient
}

type StackUpgradeOperations interface {
	List(opts *ListOpts) (*StackUpgradeCollection, error)
	Create(opts *StackUpgrade) (*StackUpgrade, error)
	Update(existing *StackUpgrade, updates interface{}) (*StackUpgrade, error)
	ById(id string) (*StackUpgrade, error)
	Delete(container *StackUpgrade) error
}

func newStackUpgradeClient(rancherClient *RancherClient) *StackUpgradeClient {
	return &StackUpgradeClient{
		rancherClient: rancherClient,
	}
}

func (c *StackUpgradeClient) Create(container *StackUpgrade) (*StackUpgrade, error) {
	resp := &StackUpgrade{}
	err := c.rancherClient.doCreate(STACK_UPGRADE_TYPE, container, resp)
	return resp, err
}

func (c *StackUpgradeClient) Update(existing *StackUpgrade, updates interface{}) (*StackUpgrade, error) {
	resp := &StackUpgrade{}
	err := c.rancherClient.doUpdate(STACK_UPGRADE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *StackUpgradeClient) List(opts *ListOpts) (*StackUpgradeCollection, error) {
	resp := &StackUpgradeCollection{}
	err := c.rancherClient.doList(STACK_UPGRADE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *StackUpgradeCollection) Next() (*StackUpgradeCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &StackUpgradeCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *StackUpgradeClient) ById(id string) (*StackUpgrade, error) {
	resp := &StackUpgrade{}
	err := c.rancherClient.doById(STACK_UPGRADE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *StackUpgradeClient) Delete(container *StackUpgrade) error {
	return c.rancherClient.doResourceDelete(STACK_UPGRADE_TYPE, &container.Resource)
}
