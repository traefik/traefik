package client

const (
	NETWORK_DRIVER_SERVICE_TYPE = "networkDriverService"
)

type NetworkDriverService struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	AssignServiceIpAddress bool `json:"assignServiceIpAddress,omitempty" yaml:"assign_service_ip_address,omitempty"`

	CreateIndex int64 `json:"createIndex,omitempty" yaml:"create_index,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	CurrentScale int64 `json:"currentScale,omitempty" yaml:"current_scale,omitempty"`

	Data map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	ExternalId string `json:"externalId,omitempty" yaml:"external_id,omitempty"`

	Fqdn string `json:"fqdn,omitempty" yaml:"fqdn,omitempty"`

	HealthState string `json:"healthState,omitempty" yaml:"health_state,omitempty"`

	InstanceIds []string `json:"instanceIds,omitempty" yaml:"instance_ids,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	LaunchConfig *LaunchConfig `json:"launchConfig,omitempty" yaml:"launch_config,omitempty"`

	LbConfig *LbTargetConfig `json:"lbConfig,omitempty" yaml:"lb_config,omitempty"`

	LinkedServices map[string]interface{} `json:"linkedServices,omitempty" yaml:"linked_services,omitempty"`

	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	NetworkDriver NetworkDriver `json:"networkDriver,omitempty" yaml:"network_driver,omitempty"`

	PublicEndpoints []PublicEndpoint `json:"publicEndpoints,omitempty" yaml:"public_endpoints,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	RetainIp bool `json:"retainIp,omitempty" yaml:"retain_ip,omitempty"`

	Scale int64 `json:"scale,omitempty" yaml:"scale,omitempty"`

	ScalePolicy *ScalePolicy `json:"scalePolicy,omitempty" yaml:"scale_policy,omitempty"`

	SecondaryLaunchConfigs []SecondaryLaunchConfig `json:"secondaryLaunchConfigs,omitempty" yaml:"secondary_launch_configs,omitempty"`

	SelectorContainer string `json:"selectorContainer,omitempty" yaml:"selector_container,omitempty"`

	SelectorLink string `json:"selectorLink,omitempty" yaml:"selector_link,omitempty"`

	StackId string `json:"stackId,omitempty" yaml:"stack_id,omitempty"`

	StartOnCreate bool `json:"startOnCreate,omitempty" yaml:"start_on_create,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	System bool `json:"system,omitempty" yaml:"system,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	TransitioningProgress int64 `json:"transitioningProgress,omitempty" yaml:"transitioning_progress,omitempty"`

	Upgrade *ServiceUpgrade `json:"upgrade,omitempty" yaml:"upgrade,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`

	Vip string `json:"vip,omitempty" yaml:"vip,omitempty"`
}

type NetworkDriverServiceCollection struct {
	Collection
	Data   []NetworkDriverService `json:"data,omitempty"`
	client *NetworkDriverServiceClient
}

type NetworkDriverServiceClient struct {
	rancherClient *RancherClient
}

type NetworkDriverServiceOperations interface {
	List(opts *ListOpts) (*NetworkDriverServiceCollection, error)
	Create(opts *NetworkDriverService) (*NetworkDriverService, error)
	Update(existing *NetworkDriverService, updates interface{}) (*NetworkDriverService, error)
	ById(id string) (*NetworkDriverService, error)
	Delete(container *NetworkDriverService) error

	ActionActivate(*NetworkDriverService) (*Service, error)

	ActionAddservicelink(*NetworkDriverService, *AddRemoveServiceLinkInput) (*Service, error)

	ActionCancelupgrade(*NetworkDriverService) (*Service, error)

	ActionContinueupgrade(*NetworkDriverService) (*Service, error)

	ActionCreate(*NetworkDriverService) (*Service, error)

	ActionDeactivate(*NetworkDriverService) (*Service, error)

	ActionFinishupgrade(*NetworkDriverService) (*Service, error)

	ActionRemove(*NetworkDriverService) (*Service, error)

	ActionRemoveservicelink(*NetworkDriverService, *AddRemoveServiceLinkInput) (*Service, error)

	ActionRestart(*NetworkDriverService, *ServiceRestart) (*Service, error)

	ActionRollback(*NetworkDriverService) (*Service, error)

	ActionSetservicelinks(*NetworkDriverService, *SetServiceLinksInput) (*Service, error)

	ActionUpdate(*NetworkDriverService) (*Service, error)

	ActionUpgrade(*NetworkDriverService, *ServiceUpgrade) (*Service, error)
}

func newNetworkDriverServiceClient(rancherClient *RancherClient) *NetworkDriverServiceClient {
	return &NetworkDriverServiceClient{
		rancherClient: rancherClient,
	}
}

func (c *NetworkDriverServiceClient) Create(container *NetworkDriverService) (*NetworkDriverService, error) {
	resp := &NetworkDriverService{}
	err := c.rancherClient.doCreate(NETWORK_DRIVER_SERVICE_TYPE, container, resp)
	return resp, err
}

func (c *NetworkDriverServiceClient) Update(existing *NetworkDriverService, updates interface{}) (*NetworkDriverService, error) {
	resp := &NetworkDriverService{}
	err := c.rancherClient.doUpdate(NETWORK_DRIVER_SERVICE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *NetworkDriverServiceClient) List(opts *ListOpts) (*NetworkDriverServiceCollection, error) {
	resp := &NetworkDriverServiceCollection{}
	err := c.rancherClient.doList(NETWORK_DRIVER_SERVICE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *NetworkDriverServiceCollection) Next() (*NetworkDriverServiceCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &NetworkDriverServiceCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *NetworkDriverServiceClient) ById(id string) (*NetworkDriverService, error) {
	resp := &NetworkDriverService{}
	err := c.rancherClient.doById(NETWORK_DRIVER_SERVICE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *NetworkDriverServiceClient) Delete(container *NetworkDriverService) error {
	return c.rancherClient.doResourceDelete(NETWORK_DRIVER_SERVICE_TYPE, &container.Resource)
}

func (c *NetworkDriverServiceClient) ActionActivate(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "activate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionAddservicelink(resource *NetworkDriverService, input *AddRemoveServiceLinkInput) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "addservicelink", &resource.Resource, input, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionCancelupgrade(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "cancelupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionContinueupgrade(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "continueupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionCreate(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionDeactivate(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "deactivate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionFinishupgrade(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "finishupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionRemove(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionRemoveservicelink(resource *NetworkDriverService, input *AddRemoveServiceLinkInput) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "removeservicelink", &resource.Resource, input, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionRestart(resource *NetworkDriverService, input *ServiceRestart) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "restart", &resource.Resource, input, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionRollback(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "rollback", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionSetservicelinks(resource *NetworkDriverService, input *SetServiceLinksInput) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "setservicelinks", &resource.Resource, input, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionUpdate(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "update", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionUpgrade(resource *NetworkDriverService, input *ServiceUpgrade) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "upgrade", &resource.Resource, input, resp)

	return resp, err
}
