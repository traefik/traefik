package egoscale

// Event represents an event in the system
type Event struct {
	Account     string `json:"account,omitempty" doc:"the account name for the account that owns the object being acted on in the event (e.g. the owner of the virtual machine, ip address, or security group)"`
	Created     string `json:"created,omitempty" doc:"the date the event was created"`
	Description string `json:"description,omitempty" doc:"a brief description of the event"`
	Domain      string `json:"domain,omitempty" doc:"the name of the account's domain"`
	DomainID    *UUID  `json:"domainid,omitempty" doc:"the id of the account's domain"`
	ID          *UUID  `json:"id,omitempty" doc:"the ID of the event"`
	Level       string `json:"level,omitempty" doc:"the event level (INFO, WARN, ERROR)"`
	ParentID    *UUID  `json:"parentid,omitempty" doc:"whether the event is parented"`
	State       string `json:"state,omitempty" doc:"the state of the event"`
	Type        string `json:"type,omitempty" doc:"the type of the event (see event types)"`
	UserName    string `json:"username,omitempty" doc:"the name of the user who performed the action (can be different from the account if an admin is performing an action for a user, e.g. starting/stopping a user's virtual machine)"`
}

// EventType represent a type of event
type EventType struct {
	Name string `json:"name,omitempty" doc:"Event Type"`
}

// ListEvents list the events
type ListEvents struct {
	Account     string `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	DomainID    *UUID  `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	Duration    int    `json:"duration,omitempty" doc:"the duration of the event"`
	EndDate     string `json:"enddate,omitempty" doc:"the end date range of the list you want to retrieve (use format \"yyyy-MM-dd\" or the new format \"yyyy-MM-dd HH:mm:ss\")"`
	EntryTime   int    `json:"entrytime,omitempty" doc:"the time the event was entered"`
	ID          *UUID  `json:"id,omitempty" doc:"the ID of the event"`
	IsRecursive *bool  `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves." doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	Keyword     string `json:"keyword,omitempty" doc:"List by keyword"`
	Level       string `json:"level,omitempty" doc:"the event level (INFO, WARN, ERROR)"`
	ListAll     *bool  `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Page        int    `json:"page,omitempty" `
	PageSize    int    `json:"pagesize,omitempty"`
	StartDate   string `json:"startdate,omitempty" doc:"the start date range of the list you want to retrieve (use format \"yyyy-MM-dd\" or the new format \"yyyy-MM-dd HH:mm:ss\")"`
	Type        string `json:"type,omitempty" doc:"the event type (see event types)"`
	_           bool   `name:"listEvents" description:"A command to list events."`
}

// ListEventsResponse represents a response of a list query
type ListEventsResponse struct {
	Count int     `json:"count"`
	Event []Event `json:"event"`
}

func (ListEvents) response() interface{} {
	return new(ListEventsResponse)
}

// ListEventTypes list the event types
type ListEventTypes struct {
	_ bool `name:"listEventTypes" description:"List Event Types"`
}

// ListEventTypesResponse represents a response of a list query
type ListEventTypesResponse struct {
	Count     int         `json:"count"`
	EventType []EventType `json:"eventtype"`
	_         bool        `name:"listEventTypes" description:"List Event Types"`
}

func (ListEventTypes) response() interface{} {
	return new(ListEventTypesResponse)
}
