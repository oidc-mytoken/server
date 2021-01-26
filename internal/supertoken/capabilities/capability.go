package capabilities

import (
	"database/sql/driver"
	"encoding/json"
)

// Constants for capabilities
const (
	CapabilityAT               Capability = "AT"
	CapabilityCreateST         Capability = "create_super_token"
	CapabilitySettings         Capability = "settings"
	CapabilityTokeninfoHistory Capability = "tokeninfo_history"
	CapabilityTokeninfoTree    Capability = "tokeninfo_tree"
	CapabilityListST           Capability = "list_super_tokens"
)

// AllCapabilities holds all defined capabilities
var AllCapabilities = Capabilities{
	CapabilityAT,
	CapabilityCreateST,
	CapabilitySettings,
	CapabilityTokeninfoHistory,
	CapabilityTokeninfoTree,
	CapabilityListST,
}

// NewCapabilities casts a []string into Capabilities
func NewCapabilities(caps []string) (c Capabilities) {
	for _, cc := range caps {
		c = append(c, Capability(cc))
	}
	return
}

// Strings returns a slice of strings for these capabilities
func (c Capabilities) Strings() (s []string) {
	for _, cc := range c {
		s = append(s, string(cc))
	}
	return
}

// Capabilities is a slice of Capability
type Capabilities []Capability

// Capability is a capability string
type Capability string

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
