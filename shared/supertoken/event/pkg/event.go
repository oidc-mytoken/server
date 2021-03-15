package event

import (
	"database/sql/driver"
	"fmt"
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
		return fmt.Errorf("unknown event")
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
	return &Event{Type: number, Comment: comment}
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
var AllEvents = [...]string{"unknown", "created", "AT_created", "ST_created", "tokeninfo_introspect", "tokeninfo_history", "tokeninfo_tree", "tokeninfo_list_super_tokens", "mng_enabled_AT_grant", "mng_disabled_AT_grant", "mng_enabled_JWT_grant", "mng_disabled_JWT_grant", "mng_linked_grant", "mng_unlinked_grant", "mng_enabled_tracing", "mng_disabled_tracing", "inherited_RT", "transfer_code_created", "transfer_code_used"}

// Events for SuperTokens
const (
	STEventUnknown = iota
	STEventCreated
	STEventATCreated
	STEventSTCreated
	STEventTokenInfoIntrospect
	STEventTokenInfoHistory
	STEventTokenInfoTree
	STEventTokenInfoListSTs
	STEventMngGrantATEnabled
	STEventMngGrantATDisabled
	STEventMngGrantJWTEnabled
	STEventMngGrantJWTDisabled
	STEventMngGrantLinked
	STEventMngGrantUnlinked
	STEventMngTracingEnabled
	STEventMngTracingDisabled
	STEventInheritedRT
	STEventTransferCodeCreated
	STEventTransferCodeUsed
	maxEvent
)
