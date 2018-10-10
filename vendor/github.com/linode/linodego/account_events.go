package linodego

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// Event represents an action taken on the Account.
type Event struct {
	CreatedStr string `json:"created"`

	// The unique ID of this Event.
	ID int `json:"id"`

	// Current status of the Event, Enum: "failed" "finished" "notification" "scheduled" "started"
	Status EventStatus `json:"status"`

	// The action that caused this Event. New actions may be added in the future.
	Action EventAction `json:"action"`

	// A percentage estimating the amount of time remaining for an Event. Returns null for notification events.
	PercentComplete int `json:"percent_complete"`

	// The rate of completion of the Event. Only some Events will return rate; for example, migration and resize Events.
	Rate *string `json:"rate"`

	// If this Event has been read.
	Read bool `json:"read"`

	// If this Event has been seen.
	Seen bool `json:"seen"`

	// The estimated time remaining until the completion of this Event. This value is only returned for in-progress events.
	TimeRemainingMsg json.RawMessage `json:"time_remaining"`
	TimeRemaining    *int            `json:"-"`

	// The username of the User who caused the Event.
	Username string `json:"username"`

	// Detailed information about the Event's entity, including ID, type, label, and URL used to access it.
	Entity *EventEntity `json:"entity"`

	// When this Event was created.
	Created *time.Time `json:"-"`
}

// EventAction constants start with Action and include all known Linode API Event Actions.
type EventAction string

// EventAction constants represent the actions that cause an Event. New actions may be added in the future.
const (
	ActionBackupsEnable            EventAction = "backups_enable"
	ActionBackupsCancel            EventAction = "backups_cancel"
	ActionBackupsRestore           EventAction = "backups_restore"
	ActionCommunityQuestionReply   EventAction = "community_question_reply"
	ActionCreateCardUpdated        EventAction = "credit_card_updated"
	ActionDiskCreate               EventAction = "disk_create"
	ActionDiskDelete               EventAction = "disk_delete"
	ActionDiskDuplicate            EventAction = "disk_duplicate"
	ActionDiskImagize              EventAction = "disk_imagize"
	ActionDiskResize               EventAction = "disk_resize"
	ActionDNSRecordCreate          EventAction = "dns_record_create"
	ActionDNSRecordDelete          EventAction = "dns_record_delete"
	ActionDNSZoneCreate            EventAction = "dns_zone_create"
	ActionDNSZoneDelete            EventAction = "dns_zone_delete"
	ActionImageDelete              EventAction = "image_delete"
	ActionLinodeAddIP              EventAction = "linode_addip"
	ActionLinodeBoot               EventAction = "linode_boot"
	ActionLinodeClone              EventAction = "linode_clone"
	ActionLinodeCreate             EventAction = "linode_create"
	ActionLinodeDelete             EventAction = "linode_delete"
	ActionLinodeDeleteIP           EventAction = "linode_deleteip"
	ActionLinodeMigrate            EventAction = "linode_migrate"
	ActionLinodeMutate             EventAction = "linode_mutate"
	ActionLinodeReboot             EventAction = "linode_reboot"
	ActionLinodeRebuild            EventAction = "linode_rebuild"
	ActionLinodeResize             EventAction = "linode_resize"
	ActionLinodeShutdown           EventAction = "linode_shutdown"
	ActionLinodeSnapshot           EventAction = "linode_snapshot"
	ActionLongviewClientCreate     EventAction = "longviewclient_create"
	ActionLongviewClientDelete     EventAction = "longviewclient_delete"
	ActionManagedDisabled          EventAction = "managed_disabled"
	ActionManagedEnabled           EventAction = "managed_enabled"
	ActionManagedServiceCreate     EventAction = "managed_service_create"
	ActionManagedServiceDelete     EventAction = "managed_service_delete"
	ActionNodebalancerCreate       EventAction = "nodebalancer_create"
	ActionNodebalancerDelete       EventAction = "nodebalancer_delete"
	ActionNodebalancerConfigCreate EventAction = "nodebalancer_config_create"
	ActionNodebalancerConfigDelete EventAction = "nodebalancer_config_delete"
	ActionPasswordReset            EventAction = "password_reset"
	ActionPaymentSubmitted         EventAction = "payment_submitted"
	ActionStackScriptCreate        EventAction = "stackscript_create"
	ActionStackScriptDelete        EventAction = "stackscript_delete"
	ActionStackScriptPublicize     EventAction = "stackscript_publicize"
	ActionStackScriptRevise        EventAction = "stackscript_revise"
	ActionTFADisabled              EventAction = "tfa_disabled"
	ActionTFAEnabled               EventAction = "tfa_enabled"
	ActionTicketAttachmentUpload   EventAction = "ticket_attachment_upload"
	ActionTicketCreate             EventAction = "ticket_create"
	ActionTicketReply              EventAction = "ticket_reply"
	ActionVolumeAttach             EventAction = "volume_attach"
	ActionVolumeClone              EventAction = "volume_clone"
	ActionVolumeCreate             EventAction = "volume_create"
	ActionVolumeDelte              EventAction = "volume_delete"
	ActionVolumeDetach             EventAction = "volume_detach"
	ActionVolumeResize             EventAction = "volume_resize"
)

// EntityType constants start with Entity and include Linode API Event Entity Types
type EntityType string

// EntityType contants are the entities an Event can be related to
const (
	EntityLinode EntityType = "linode"
	EntityDisk   EntityType = "disk"
)

// EventStatus constants start with Event and include Linode API Event Status values
type EventStatus string

// EventStatus constants reflect the current status of an Event
const (
	EventFailed       EventStatus = "failed"
	EventFinished     EventStatus = "finished"
	EventNotification EventStatus = "notification"
	EventScheduled    EventStatus = "scheduled"
	EventStarted      EventStatus = "started"
)

// EventEntity provides detailed information about the Event's
// associated entity, including ID, Type, Label, and a URL that
// can be used to access it.
type EventEntity struct {
	// ID may be a string or int, it depends on the EntityType
	ID    interface{} `json:"id"`
	Label string      `json:"label"`
	Type  EntityType  `json:"type"`
	URL   string      `json:"url"`
}

// EventsPagedResponse represents a paginated Events API response
type EventsPagedResponse struct {
	*PageOptions
	Data []Event `json:"data"`
}

// endpoint gets the endpoint URL for Event
func (EventsPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.Events.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

// endpointWithID gets the endpoint URL for a specific Event
func (e Event) endpointWithID(c *Client) string {
	endpoint, err := c.Events.Endpoint()
	if err != nil {
		panic(err)
	}
	endpoint = fmt.Sprintf("%s/%d", endpoint, e.ID)
	return endpoint
}

// appendData appends Events when processing paginated Event responses
func (resp *EventsPagedResponse) appendData(r *EventsPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListEvents gets a collection of Event objects representing actions taken
// on the Account. The Events returned depend on the token grants and the grants
// of the associated user.
func (c *Client) ListEvents(ctx context.Context, opts *ListOptions) ([]Event, error) {
	response := EventsPagedResponse{}
	err := c.listHelper(ctx, &response, opts)
	for i := range response.Data {
		response.Data[i].fixDates()
	}
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

// GetEvent gets the Event with the Event ID
func (c *Client) GetEvent(ctx context.Context, id int) (*Event, error) {
	e, err := c.Events.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, id)
	r, err := c.R(ctx).SetResult(&Event{}).Get(e)
	if err != nil {
		return nil, err
	}
	return r.Result().(*Event).fixDates(), nil
}

// fixDates converts JSON timestamps to Go time.Time values
func (e *Event) fixDates() *Event {
	e.Created, _ = parseDates(e.CreatedStr)
	e.TimeRemaining = unmarshalTimeRemaining(e.TimeRemainingMsg)
	return e
}

// MarkEventRead marks a single Event as read.
func (c *Client) MarkEventRead(ctx context.Context, event *Event) error {
	e := event.endpointWithID(c)
	e = fmt.Sprintf("%s/read", e)

	_, err := coupleAPIErrors(c.R(ctx).Post(e))

	return err
}

// MarkEventsSeen marks all Events up to and including this Event by ID as seen.
func (c *Client) MarkEventsSeen(ctx context.Context, event *Event) error {
	e := event.endpointWithID(c)
	e = fmt.Sprintf("%s/seen", e)

	_, err := coupleAPIErrors(c.R(ctx).Post(e))

	return err
}

func unmarshalTimeRemaining(m json.RawMessage) *int {
	jsonBytes, err := m.MarshalJSON()
	if err != nil {
		panic(jsonBytes)
	}

	if len(jsonBytes) == 4 && string(jsonBytes) == "null" {
		return nil
	}

	var timeStr string
	if err := json.Unmarshal(jsonBytes, &timeStr); err == nil && len(timeStr) > 0 {
		if dur, err := durationToSeconds(timeStr); err != nil {
			panic(err)
		} else {
			return &dur
		}
	} else {
		var intPtr int
		if err := json.Unmarshal(jsonBytes, &intPtr); err == nil {
			return &intPtr
		}
	}

	log.Println("[WARN] Unexpected unmarshalTimeRemaining value: ", jsonBytes)
	return nil
}

// durationToSeconds takes a hh:mm:ss string and returns the number of seconds
func durationToSeconds(s string) (int, error) {
	multipliers := [3]int{60 * 60, 60, 1}
	segs := strings.Split(s, ":")
	if len(segs) > len(multipliers) {
		return 0, fmt.Errorf("too many ':' separators in time duration: %s", s)
	}
	var d int
	l := len(segs)
	for i := 0; i < l; i++ {
		m, err := strconv.Atoi(segs[i])
		if err != nil {
			return 0, err
		}
		d += m * multipliers[i+len(multipliers)-l]
	}
	return d, nil
}
