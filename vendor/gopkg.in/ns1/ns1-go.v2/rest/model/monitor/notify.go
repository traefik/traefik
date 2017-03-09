package monitor

// NotifyList wraps notifications.
type NotifyList struct {
	ID            string          `json:"id,omitempty"`
	Name          string          `json:"name,omitempty"`
	Notifications []*Notification `json:"notify_list,omitempty"`
}

// Notification represents endpoint to alert to.
type Notification struct {
	Type   string `json:"type,omitempty"`
	Config Config `json:"config,omitempty"`
}

// NewNotifyList returns a notify list that alerts via the given notifications.
func NewNotifyList(name string, nl ...*Notification) *NotifyList {
	if nl == nil {
		nl = []*Notification{}
	}

	return &NotifyList{Name: name, Notifications: nl}
}

// NewUserNotification returns a notification that alerts via user.
func NewUserNotification(username string) *Notification {
	return &Notification{
		Type:   "user",
		Config: Config{"user": username}}
}

// NewEmailNotification returns a notification that alerts via email.
func NewEmailNotification(email string) *Notification {
	return &Notification{
		Type:   "email",
		Config: Config{"email": email}}
}

// NewFeedNotification returns a notification that alerts via datafeed.
func NewFeedNotification(sourceID string) *Notification {
	return &Notification{
		Type:   "datafeed",
		Config: Config{"sourceid": sourceID}}
}

// NewWebNotification returns a notification that alerts via webhook.
func NewWebNotification(url string) *Notification {
	return &Notification{
		Type:   "webhook",
		Config: Config{"url": url}}
}

// NewPagerDutyNotification returns a notification that alerts via pagerduty.
func NewPagerDutyNotification(key string) *Notification {
	return &Notification{
		Type:   "pagerduty",
		Config: Config{"service_key": key}}
}

// NewHipChatNotification returns a notification that alerts via hipchat.
func NewHipChatNotification(token, room string) *Notification {
	return &Notification{
		Type:   "hipchat",
		Config: Config{"token": token, "room": room}}
}

// NewSlackNotification returns a notification that alerts via slack.
func NewSlackNotification(url, username, channel string) *Notification {
	return &Notification{
		Type:   "slack",
		Config: Config{"url": url, "username": username, "channel": channel}}
}
