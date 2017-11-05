package client

const (
	SERVICE_BINDING_TYPE = "serviceBinding"
)

type ServiceBinding struct {
	Resource

	Labels map[string]interface{} `json:"labels,omitempty" yaml:"labels,omitempty"`

	Ports []string `json:"ports,omitempty" yaml:"ports,omitempty"`
}

type ServiceBindingCollection struct {
	Collection
	Data   []ServiceBinding `json:"data,omitempty"`
	client *ServiceBindingClient
}

type ServiceBindingClient struct {
	rancherClient *RancherClient
}

type ServiceBindingOperations interface {
	List(opts *ListOpts) (*ServiceBindingCollection, error)
	Create(opts *ServiceBinding) (*ServiceBinding, error)
	Update(existing *ServiceBinding, updates interface{}) (*ServiceBinding, error)
	ById(id string) (*ServiceBinding, error)
	Delete(container *ServiceBinding) error
}

func newServiceBindingClient(rancherClient *RancherClient) *ServiceBindingClient {
	return &ServiceBindingClient{
		rancherClient: rancherClient,
	}
}

func (c *ServiceBindingClient) Create(container *ServiceBinding) (*ServiceBinding, error) {
	resp := &ServiceBinding{}
	err := c.rancherClient.doCreate(SERVICE_BINDING_TYPE, container, resp)
	return resp, err
}

func (c *ServiceBindingClient) Update(existing *ServiceBinding, updates interface{}) (*ServiceBinding, error) {
	resp := &ServiceBinding{}
	err := c.rancherClient.doUpdate(SERVICE_BINDING_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ServiceBindingClient) List(opts *ListOpts) (*ServiceBindingCollection, error) {
	resp := &ServiceBindingCollection{}
	err := c.rancherClient.doList(SERVICE_BINDING_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ServiceBindingCollection) Next() (*ServiceBindingCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ServiceBindingCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ServiceBindingClient) ById(id string) (*ServiceBinding, error) {
	resp := &ServiceBinding{}
	err := c.rancherClient.doById(SERVICE_BINDING_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *ServiceBindingClient) Delete(container *ServiceBinding) error {
	return c.rancherClient.doResourceDelete(SERVICE_BINDING_TYPE, &container.Resource)
}
