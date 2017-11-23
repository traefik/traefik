package client

const (
	ULIMIT_TYPE = "ulimit"
)

type Ulimit struct {
	Resource

	Hard int64 `json:"hard,omitempty" yaml:"hard,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	Soft int64 `json:"soft,omitempty" yaml:"soft,omitempty"`
}

type UlimitCollection struct {
	Collection
	Data   []Ulimit `json:"data,omitempty"`
	client *UlimitClient
}

type UlimitClient struct {
	rancherClient *RancherClient
}

type UlimitOperations interface {
	List(opts *ListOpts) (*UlimitCollection, error)
	Create(opts *Ulimit) (*Ulimit, error)
	Update(existing *Ulimit, updates interface{}) (*Ulimit, error)
	ById(id string) (*Ulimit, error)
	Delete(container *Ulimit) error
}

func newUlimitClient(rancherClient *RancherClient) *UlimitClient {
	return &UlimitClient{
		rancherClient: rancherClient,
	}
}

func (c *UlimitClient) Create(container *Ulimit) (*Ulimit, error) {
	resp := &Ulimit{}
	err := c.rancherClient.doCreate(ULIMIT_TYPE, container, resp)
	return resp, err
}

func (c *UlimitClient) Update(existing *Ulimit, updates interface{}) (*Ulimit, error) {
	resp := &Ulimit{}
	err := c.rancherClient.doUpdate(ULIMIT_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *UlimitClient) List(opts *ListOpts) (*UlimitCollection, error) {
	resp := &UlimitCollection{}
	err := c.rancherClient.doList(ULIMIT_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *UlimitCollection) Next() (*UlimitCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &UlimitCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *UlimitClient) ById(id string) (*Ulimit, error) {
	resp := &Ulimit{}
	err := c.rancherClient.doById(ULIMIT_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *UlimitClient) Delete(container *Ulimit) error {
	return c.rancherClient.doResourceDelete(ULIMIT_TYPE, &container.Resource)
}
