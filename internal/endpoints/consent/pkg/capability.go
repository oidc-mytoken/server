package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/shared/utils"
)

// WebCapability is type for representing api.Capability in the consent screen
type WebCapability struct {
	api.Capability
	intClass *int
}

// WebCapabilities creates a slice of WebCapability from api.Capabilities
func WebCapabilities(cc api.Capabilities) (wc []WebCapability) {
	for _, c := range cc {
		wc = append(wc, WebCapability{c, nil})
	}
	return
}

// internal classes
const (
	intClassNormal = iota
	intClassWarning
	intClassDanger
)

var normalCapabilities = []string{
	api.CapabilityAT.Name,
	api.CapabilityCreateMT.Name,
	api.CapabilityTokeninfoIntrospect.Name,
	api.CapabilityTokeninfoHistory.Name,
	api.CapabilityTokeninfoTree.Name,
}
var warningCapabilities = []string{api.CapabilityListMT.Name}
var dangerCapabilities = []string{api.CapabilitySettings.Name}

func (c WebCapability) getIntClass() int {
	if c.intClass != nil {
		return *c.intClass
	}
	if utils.StringInSlice(c.Name, normalCapabilities) {
		c.intClass = utils.NewInt(intClassNormal)
	}
	if utils.StringInSlice(c.Name, warningCapabilities) {
		c.intClass = utils.NewInt(intClassWarning)
	}
	if utils.StringInSlice(c.Name, dangerCapabilities) {
		c.intClass = utils.NewInt(intClassDanger)
	}
	if c.intClass != nil {
		return *c.intClass
	}
	return -1
}

func (c WebCapability) getDangerLevel() int {
	return c.getIntClass()
}

// ColorClass returns the html class for coloring this Capability
func (c WebCapability) ColorClass() string {
	return textColorByDanger(c.getDangerLevel())
}

// CapabilityLevel returns a string describing the power of this capability
func (c WebCapability) CapabilityLevel() string {
	intClass := c.getIntClass()
	switch intClass {
	case 0:
		return "This is a normal capability."
	case 1:
		return "This is a powerful capability!"
	case 2:
		return "This is a very powerful capability!"
	}
	return ""
}

// IsCreateMT checks if this WebCapability is api.CapabilityCreateMT
func (c WebCapability) IsCreateMT() bool {
	return c.Name == api.CapabilityCreateMT.Name
}
