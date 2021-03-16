package pkg

import (
	"github.com/oidc-mytoken/server/shared/mytoken/capabilities"
	"github.com/oidc-mytoken/server/shared/utils"
)

type WebCapability struct {
	capabilities.Capability
	intClass *int
}

func WebCapabilities(cc capabilities.Capabilities) (wc []WebCapability) {
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
	capabilities.CapabilityAT.Name,
	capabilities.CapabilityCreateMT.Name,
	capabilities.CapabilityTokeninfoIntrospect.Name,
	capabilities.CapabilityTokeninfoHistory.Name,
	capabilities.CapabilityTokeninfoTree.Name,
}
var warningCapabilities = []string{capabilities.CapabilityListMT.Name}
var dangerCapabilities = []string{capabilities.CapabilitySettings.Name}

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

func (c WebCapability) ColorClass() string {
	return textColorByDanger(c.getDangerLevel())
}

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

func (c WebCapability) IsCreateMT() bool {
	return c.Name == capabilities.CapabilityCreateMT.Name
}
