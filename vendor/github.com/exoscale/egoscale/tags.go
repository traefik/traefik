package egoscale

// ResourceTag is a tag associated with a resource
//
// http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/4.9/management.html
type ResourceTag struct {
	Account      string `json:"account,omitempty" doc:"the account associated with the tag"`
	Customer     string `json:"customer,omitempty" doc:"customer associated with the tag"`
	Domain       string `json:"domain,omitempty" doc:"the domain associated with the tag"`
	DomainID     *UUID  `json:"domainid,omitempty" doc:"the ID of the domain associated with the tag"`
	Key          string `json:"key,omitempty" doc:"tag key name"`
	ResourceID   *UUID  `json:"resourceid,omitempty" doc:"id of the resource"`
	ResourceType string `json:"resourcetype,omitempty" doc:"resource type"`
	Value        string `json:"value,omitempty" doc:"tag value"`
}

// CreateTags (Async) creates resource tag(s)
type CreateTags struct {
	ResourceIDs  []UUID        `json:"resourceids" doc:"list of resources to create the tags for"`
	ResourceType string        `json:"resourcetype" doc:"type of the resource"`
	Tags         []ResourceTag `json:"tags" doc:"Map of tags (key/value pairs)"`
	Customer     string        `json:"customer,omitempty" doc:"identifies client specific tag. When the value is not null, the tag can't be used by cloudStack code internally"`
	_            bool          `name:"createTags" description:"Creates resource tag(s)"`
}

func (CreateTags) response() interface{} {
	return new(AsyncJobResult)
}

func (CreateTags) asyncResponse() interface{} {
	return new(booleanResponse)
}

// DeleteTags (Async) deletes the resource tag(s)
type DeleteTags struct {
	ResourceIDs  []UUID        `json:"resourceids" doc:"Delete tags for resource id(s)"`
	ResourceType string        `json:"resourcetype" doc:"Delete tag by resource type"`
	Tags         []ResourceTag `json:"tags,omitempty" doc:"Delete tags matching key/value pairs"`
	_            bool          `name:"deleteTags" description:"Deleting resource tag(s)"`
}

func (DeleteTags) response() interface{} {
	return new(AsyncJobResult)
}

func (DeleteTags) asyncResponse() interface{} {
	return new(booleanResponse)
}

// ListTags list resource tag(s)
type ListTags struct {
	Account      string `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	Customer     string `json:"customer,omitempty" doc:"list by customer name"`
	DomainID     *UUID  `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	IsRecursive  *bool  `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	Key          string `json:"key,omitempty" doc:"list by key"`
	Keyword      string `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll      *bool  `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
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

func (ListTags) response() interface{} {
	return new(ListTagsResponse)
}
