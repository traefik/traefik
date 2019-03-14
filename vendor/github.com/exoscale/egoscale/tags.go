package egoscale

// ResourceTag is a tag associated with a resource
//
// https://community.exoscale.com/documentation/compute/instance-tags/
type ResourceTag struct {
	Account      string `json:"account,omitempty" doc:"the account associated with the tag"`
	Customer     string `json:"customer,omitempty" doc:"customer associated with the tag"`
	Key          string `json:"key,omitempty" doc:"tag key name"`
	ResourceID   *UUID  `json:"resourceid,omitempty" doc:"id of the resource"`
	ResourceType string `json:"resourcetype,omitempty" doc:"resource type"`
	Value        string `json:"value,omitempty" doc:"tag value"`
}

// ListRequest builds the ListZones request
func (tag ResourceTag) ListRequest() (ListCommand, error) {
	req := &ListTags{
		Customer:     tag.Customer,
		Key:          tag.Key,
		ResourceID:   tag.ResourceID,
		ResourceType: tag.ResourceType,
		Value:        tag.Value,
	}

	return req, nil
}

// CreateTags (Async) creates resource tag(s)
type CreateTags struct {
	ResourceIDs  []UUID        `json:"resourceids" doc:"list of resources to create the tags for"`
	ResourceType string        `json:"resourcetype" doc:"type of the resource"`
	Tags         []ResourceTag `json:"tags" doc:"Map of tags (key/value pairs)"`
	Customer     string        `json:"customer,omitempty" doc:"identifies client specific tag. When the value is not null, the tag can't be used by cloudStack code internally"`
	_            bool          `name:"createTags" description:"Creates resource tag(s)"`
}

// Response returns the struct to unmarshal
func (CreateTags) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (CreateTags) AsyncResponse() interface{} {
	return new(BooleanResponse)
}

// DeleteTags (Async) deletes the resource tag(s)
type DeleteTags struct {
	ResourceIDs  []UUID        `json:"resourceids" doc:"Delete tags for resource id(s)"`
	ResourceType string        `json:"resourcetype" doc:"Delete tag by resource type"`
	Tags         []ResourceTag `json:"tags,omitempty" doc:"Delete tags matching key/value pairs"`
	_            bool          `name:"deleteTags" description:"Deleting resource tag(s)"`
}

// Response returns the struct to unmarshal
func (DeleteTags) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (DeleteTags) AsyncResponse() interface{} {
	return new(BooleanResponse)
}

//go:generate go run generate/main.go -interface=Listable ListTags

// ListTags list resource tag(s)
type ListTags struct {
	Customer     string `json:"customer,omitempty" doc:"list by customer name"`
	Key          string `json:"key,omitempty" doc:"list by key"`
	Keyword      string `json:"keyword,omitempty" doc:"List by keyword"`
	Page         int    `json:"page,omitempty"`
	PageSize     int    `json:"pagesize,omitempty"`
	ResourceID   *UUID  `json:"resourceid,omitempty" doc:"list by resource id"`
	ResourceType string `json:"resourcetype,omitempty" doc:"list by resource type"`
	Value        string `json:"value,omitempty" doc:"list by value"`
	_            bool   `name:"listTags" description:"List resource tag(s)"`
}

// ListTagsResponse represents a list of resource tags
type ListTagsResponse struct {
	Count int           `json:"count"`
	Tag   []ResourceTag `json:"tag"`
}
