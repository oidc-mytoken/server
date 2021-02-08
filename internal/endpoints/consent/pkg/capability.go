package pkg

import (
	"github.com/oidc-mytoken/server/shared/supertoken/capabilities"
	"github.com/oidc-mytoken/server/shared/utils"
)

type webCapability struct {
	capabilities.Capability
	intClass *int
}

func WebCapabilities(cc capabilities.Capabilities) (wc []webCapability) {
	for _, c := range cc {
		wc = append(wc, webCapability{c, nil})
	}
	return
}

// internal classes
const (
	intClassNormal = iota
	intClassWarning
	intClassDanger
)

var normalCapabilities = []string{"AT", "create_super_token", "tokeninfo_introspect", "tokeninfo_history", "tokeninfo_tree"}
var warningCapabilities = []string{"list_super_tokens"}
var dangerCapabilities = []string{"settings"}

func (c webCapability) getIntClass() int {
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

func (c webCapability) getDangerLevel() int {
	return c.getIntClass()
}

func (c webCapability) ColorClass() string {
	return textColorByDanger(c.getDangerLevel())
}

func (c webCapability) CapabilityLevel() string {
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

func (c webCapability) IsCreateST() bool {
	return c.Name == capabilities.CapabilityCreateST.Name
}
