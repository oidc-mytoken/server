package event

import (
	"database/sql/driver"

	"github.com/pkg/errors"
)

// Event is an enum like type for events
type Event struct {
	Type    int
	Comment string
}

func (e *Event) String() string {
	if e.Type < 0 || e.Type >= len(AllEvents) {
		return ""
	}
	return AllEvents[e.Type]
}

// Valid checks that Event is a defined Event
func (e *Event) Valid() bool {
	return e.Type < maxEvent && e.Type >= 0
}

// Value implements the sql.Valuer interface
func (e *Event) Value() (driver.Value, error) {
	return e.String(), nil
}

// Scan implements the sql.Scanner interface
func (e *Event) Scan(src interface{}) error {
	number := eventStringToInt(src.(string))
	if number < 0 {
		return errors.New("unknown event")
	}
	e.Type = number
	return nil
}

// NewEvent creates a new Event from the event string
func NewEvent(typ, comment string) *Event {
	number := eventStringToInt(typ)
	return FromNumber(number, comment)
}

// FromNumber creates an Event from the number
func FromNumber(number int, comment string) *Event {
	if number < 0 {
		return nil
	}
	return &Event{
		Type:    number,
		Comment: comment,
	}
}

func eventStringToInt(str string) int {
	for i, e := range AllEvents {
		if str == e {
			return i
		}
	}
	return -1
}

// AllEvents hold all possible Events
var AllEvents = [...]string{
	"unknown",
	"created",
	"AT_created",
	"MT_created",
	"tokeninfo_introspect",
	"tokeninfo_history",
	"tokeninfo_subtokens",
	"tokeninfo_list_mytokens",
	"inherited_RT",
	"transfer_code_created",
	"transfer_code_used",
	"token_rotated",
	"settings_grant_enabled",
	"settings_grant_disabled",
	"settings_grants_listed",
	"ssh_keys_listed",
	"ssh_key_added",
	"revoked_other_token",
	"tokeninfo_history_other_token",
	"expired",
	"revoked",
	"notification_subscribed",
	"notification_listed",
	"notification_unsubscribed",
	"notification_subscribed_other",
	"notification_unsubscribed_other",
	"calendar_created",
	"calendar_listed",
	"calendar_deleted",
}

// Events for Mytokens
const (
	UnknownEvent = iota
	MTCreated
	ATCreated
	SubtokenCreated
	TokenInfoIntrospect
	TokenInfoHistory
	TokenInfoSubtokens
	TokenInfoListMTs
	InheritedRT
	TransferCodeCreated
	TransferCodeUsed
	MTRotated
	GrantEnabled
	GrantDisabled
	GrantsListed
	SSHKeyListed
	SSHKeyAdded
	RevokedOtherToken
	TokenInfoHistoryOtherToken
	Expired
	Revoked
	NotificationSubscribed
	NotificationListed
	NotificationUnsubscribed
	NotificationSubscribedOther
	NotificationUnsubscribedOther
	CalendarCreated
	CalendarListed
	CalendarDeleted
	maxEvent
)
