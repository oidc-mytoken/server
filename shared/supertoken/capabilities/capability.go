package capabilities

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
)

// Defined Capabilities
var (
	CapabilityAT = Capability{
		Name:        "AT",
		Description: "Allows obtaining OpenID Connect Access Tokens.",
	}
	CapabilityCreateST = Capability{
		Name:        "create_super_token",
		Description: "Allows to create a new Super Token.",
	}
	CapabilitySettings = Capability{
		Name:        "settings",
		Description: "Allows to modify user settings.",
	}
	CapabilityTokeninfoIntrospect = Capability{
		Name:        "tokeninfo_introspect",
		Description: "Allows to obtain basic information about this token.",
	}
	CapabilityTokeninfoHistory = Capability{
		Name:        "tokeninfo_history",
		Description: "Allows to obtain the event history for this token.",
	}
	CapabilityTokeninfoTree = Capability{
		Name:        "tokeninfo_tree",
		Description: "Allows to list a subtoken-tree for this token.",
	}
	CapabilityListST = Capability{
		Name:        "list_super_tokens",
		Description: "Allows to list all super tokens.",
	}
)

// AllCapabilities holds all defined Capabilities
var AllCapabilities = Capabilities{
	CapabilityAT,
	CapabilityCreateST,
	CapabilitySettings,
	CapabilityTokeninfoIntrospect,
	CapabilityTokeninfoHistory,
	CapabilityTokeninfoTree,
	CapabilityListST,
}

func descriptionFor(name string) string {
	for _, c := range AllCapabilities {
		if strings.ToLower(c.Name) == strings.ToLower(name) {
			return c.Description
		}
	}
	return ""
}

// NewCapabilities casts a []string into Capabilities
func NewCapabilities(caps []string) (c Capabilities) {
	for _, cc := range caps {
		c = append(c, NewCapability(cc))
	}
	return
}

// NewCapability casts a string into a Capability
func NewCapability(name string) Capability {
	return Capability{
		Name:        name,
		Description: descriptionFor(name),
	}
}

// Strings returns a slice of strings for these capabilities
func (c Capabilities) Strings() (s []string) {
	for _, cc := range c {
		s = append(s, cc.Name)
	}
	return
}

// Capabilities is a slice of Capability
type Capabilities []Capability

// Capability is a capability string
type Capability struct {
	Name        string
	Description string
}

// MarshalJSON implements the json.Marshaler interface
func (c Capability) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Name)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (c *Capability) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	*c = NewCapability(name)
	return nil
}

// Scan implements the sql.Scanner interface.
func (c *Capabilities) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	val := src.([]uint8)
	err := json.Unmarshal(val, &c)
	return err
}

// Value implements the driver.Valuer interface
func (c Capabilities) Value() (driver.Value, error) {
	if len(c) == 0 {
		return nil, nil
	}
	return json.Marshal(c)
}

// Tighten tightens two set of Capabilities into one new
func Tighten(a, b Capabilities) (res Capabilities) {
	if b == nil {
		return a
	}
	for _, bb := range b {
		if a.Has(bb) {
			res = append(res, bb)
		}
	}
	return
}

// Has checks if Capabilities slice contains the passed Capability
func (c Capabilities) Has(a Capability) bool {
	for _, cc := range c {
		if cc == a {
			return true
		}
	}
	return false
}
