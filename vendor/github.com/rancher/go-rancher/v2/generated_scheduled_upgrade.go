package client

const (
	SCHEDULED_UPGRADE_TYPE = "scheduledUpgrade"
)

type ScheduledUpgrade struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	Data map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	StackId string `json:"stackId,omitempty" yaml:"stack_id,omitempty"`

	Started string `json:"started,omitempty" yaml:"started,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	TransitioningProgress int64 `json:"transitioningProgress,omitempty" yaml:"transitioning_progress,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type ScheduledUpgradeCollection struct {
	Collection
	Data   []ScheduledUpgrade `json:"data,omitempty"`
	client *ScheduledUpgradeClient
}

type ScheduledUpgradeClient struct {
	rancherClient *RancherClient
}

type ScheduledUpgradeOperations interface {
	List(opts *ListOpts) (*ScheduledUpgradeCollection, error)
	Create(opts *ScheduledUpgrade) (*ScheduledUpgrade, error)
	Update(existing *ScheduledUpgrade, updates interface{}) (*ScheduledUpgrade, error)
	ById(id string) (*ScheduledUpgrade, error)
	Delete(container *ScheduledUpgrade) error

	ActionCreate(*ScheduledUpgrade) (*ScheduledUpgrade, error)

	ActionRemove(*ScheduledUpgrade) (*ScheduledUpgrade, error)

	ActionStart(*ScheduledUpgrade) (*ScheduledUpgrade, error)
}

func newScheduledUpgradeClient(rancherClient *RancherClient) *ScheduledUpgradeClient {
	return &ScheduledUpgradeClient{
		rancherClient: rancherClient,
	}
}

func (c *ScheduledUpgradeClient) Create(container *ScheduledUpgrade) (*ScheduledUpgrade, error) {
	resp := &ScheduledUpgrade{}
	err := c.rancherClient.doCreate(SCHEDULED_UPGRADE_TYPE, container, resp)
	return resp, err
}

func (c *ScheduledUpgradeClient) Update(existing *ScheduledUpgrade, updates interface{}) (*ScheduledUpgrade, error) {
	resp := &ScheduledUpgrade{}
	err := c.rancherClient.doUpdate(SCHEDULED_UPGRADE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ScheduledUpgradeClient) List(opts *ListOpts) (*ScheduledUpgradeCollection, error) {
	resp := &ScheduledUpgradeCollection{}
	err := c.rancherClient.doList(SCHEDULED_UPGRADE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ScheduledUpgradeCollection) Next() (*ScheduledUpgradeCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ScheduledUpgradeCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ScheduledUpgradeClient) ById(id string) (*ScheduledUpgrade, error) {
	resp := &ScheduledUpgrade{}
	err := c.rancherClient.doById(SCHEDULED_UPGRADE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *ScheduledUpgradeClient) Delete(container *ScheduledUpgrade) error {
	return c.rancherClient.doResourceDelete(SCHEDULED_UPGRADE_TYPE, &container.Resource)
}

func (c *ScheduledUpgradeClient) ActionCreate(resource *ScheduledUpgrade) (*ScheduledUpgrade, error) {

	resp := &ScheduledUpgrade{}

	err := c.rancherClient.doAction(SCHEDULED_UPGRADE_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ScheduledUpgradeClient) ActionRemove(resource *ScheduledUpgrade) (*ScheduledUpgrade, error) {

	resp := &ScheduledUpgrade{}

	err := c.rancherClient.doAction(SCHEDULED_UPGRADE_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ScheduledUpgradeClient) ActionStart(resource *ScheduledUpgrade) (*ScheduledUpgrade, error) {

	resp := &ScheduledUpgrade{}

	err := c.rancherClient.doAction(SCHEDULED_UPGRADE_TYPE, "start", &resource.Resource, nil, resp)

	return resp, err
}
