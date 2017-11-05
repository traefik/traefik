package client

const (
	CHANGE_SECRET_INPUT_TYPE = "changeSecretInput"
)

type ChangeSecretInput struct {
	Resource

	NewSecret string `json:"newSecret,omitempty" yaml:"new_secret,omitempty"`

	OldSecret string `json:"oldSecret,omitempty" yaml:"old_secret,omitempty"`
}

type ChangeSecretInputCollection struct {
	Collection
	Data   []ChangeSecretInput `json:"data,omitempty"`
	client *ChangeSecretInputClient
}

type ChangeSecretInputClient struct {
	rancherClient *RancherClient
}

type ChangeSecretInputOperations interface {
	List(opts *ListOpts) (*ChangeSecretInputCollection, error)
	Create(opts *ChangeSecretInput) (*ChangeSecretInput, error)
	Update(existing *ChangeSecretInput, updates interface{}) (*ChangeSecretInput, error)
	ById(id string) (*ChangeSecretInput, error)
	Delete(container *ChangeSecretInput) error
}

func newChangeSecretInputClient(rancherClient *RancherClient) *ChangeSecretInputClient {
	return &ChangeSecretInputClient{
		rancherClient: rancherClient,
	}
}

func (c *ChangeSecretInputClient) Create(container *ChangeSecretInput) (*ChangeSecretInput, error) {
	resp := &ChangeSecretInput{}
	err := c.rancherClient.doCreate(CHANGE_SECRET_INPUT_TYPE, container, resp)
	return resp, err
}

func (c *ChangeSecretInputClient) Update(existing *ChangeSecretInput, updates interface{}) (*ChangeSecretInput, error) {
	resp := &ChangeSecretInput{}
	err := c.rancherClient.doUpdate(CHANGE_SECRET_INPUT_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ChangeSecretInputClient) List(opts *ListOpts) (*ChangeSecretInputCollection, error) {
	resp := &ChangeSecretInputCollection{}
	err := c.rancherClient.doList(CHANGE_SECRET_INPUT_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ChangeSecretInputCollection) Next() (*ChangeSecretInputCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ChangeSecretInputCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ChangeSecretInputClient) ById(id string) (*ChangeSecretInput, error) {
	resp := &ChangeSecretInput{}
	err := c.rancherClient.doById(CHANGE_SECRET_INPUT_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *ChangeSecretInputClient) Delete(container *ChangeSecretInput) error {
	return c.rancherClient.doResourceDelete(CHANGE_SECRET_INPUT_TYPE, &container.Resource)
}
