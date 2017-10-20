package client

const (
	PROCESS_POOL_TYPE = "processPool"
)

type ProcessPool struct {
	Resource

	ActiveTasks int64 `json:"activeTasks,omitempty" yaml:"active_tasks,omitempty"`

	CompletedTasks int64 `json:"completedTasks,omitempty" yaml:"completed_tasks,omitempty"`

	MaxPoolSize int64 `json:"maxPoolSize,omitempty" yaml:"max_pool_size,omitempty"`

	MinPoolSize int64 `json:"minPoolSize,omitempty" yaml:"min_pool_size,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	PoolSize int64 `json:"poolSize,omitempty" yaml:"pool_size,omitempty"`

	QueueRemainingCapacity int64 `json:"queueRemainingCapacity,omitempty" yaml:"queue_remaining_capacity,omitempty"`

	QueueSize int64 `json:"queueSize,omitempty" yaml:"queue_size,omitempty"`

	RejectedTasks int64 `json:"rejectedTasks,omitempty" yaml:"rejected_tasks,omitempty"`
}

type ProcessPoolCollection struct {
	Collection
	Data   []ProcessPool `json:"data,omitempty"`
	client *ProcessPoolClient
}

type ProcessPoolClient struct {
	rancherClient *RancherClient
}

type ProcessPoolOperations interface {
	List(opts *ListOpts) (*ProcessPoolCollection, error)
	Create(opts *ProcessPool) (*ProcessPool, error)
	Update(existing *ProcessPool, updates interface{}) (*ProcessPool, error)
	ById(id string) (*ProcessPool, error)
	Delete(container *ProcessPool) error
}

func newProcessPoolClient(rancherClient *RancherClient) *ProcessPoolClient {
	return &ProcessPoolClient{
		rancherClient: rancherClient,
	}
}

func (c *ProcessPoolClient) Create(container *ProcessPool) (*ProcessPool, error) {
	resp := &ProcessPool{}
	err := c.rancherClient.doCreate(PROCESS_POOL_TYPE, container, resp)
	return resp, err
}

func (c *ProcessPoolClient) Update(existing *ProcessPool, updates interface{}) (*ProcessPool, error) {
	resp := &ProcessPool{}
	err := c.rancherClient.doUpdate(PROCESS_POOL_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ProcessPoolClient) List(opts *ListOpts) (*ProcessPoolCollection, error) {
	resp := &ProcessPoolCollection{}
	err := c.rancherClient.doList(PROCESS_POOL_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ProcessPoolCollection) Next() (*ProcessPoolCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ProcessPoolCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ProcessPoolClient) ById(id string) (*ProcessPool, error) {
	resp := &ProcessPool{}
	err := c.rancherClient.doById(PROCESS_POOL_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *ProcessPoolClient) Delete(container *ProcessPool) error {
	return c.rancherClient.doResourceDelete(PROCESS_POOL_TYPE, &container.Resource)
}
