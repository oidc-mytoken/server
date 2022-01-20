package pkg

import (
	"strings"

	"github.com/jinzhu/copier"
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/shared/utils"
)

// WebCapability is type for representing api.Capability in the consent screen
type WebCapability struct {
	ReadWriteCapability webCapability
	ReadOnlyCapability  *webCapability
	Children            []*WebCapability
}

type webCapability struct {
	api.Capability
	intClass   *int
	IsReadOnly bool
}

// WebCapabilities creates a slice of WebCapability from api.Capabilities
func WebCapabilities(cc api.Capabilities) (wc []*WebCapability) {
	for _, c := range cc {
		wc = append(
			wc, webCapabilityFromCapability(c),
		)
	}
	return
}

// AllWebCapabilities returns all WebCapabilities as a tree
func AllWebCapabilities() []*WebCapability {
	return allWebCapabilities
}

var allWebCapabilities []*WebCapability

func init() {
	if allWebCapabilities == nil {
		allWebCapabilities = []*WebCapability{}
	}
	capabilitiesByLevel := make(map[int][]webCapability)
	var maxLevel int
	for _, c := range api.AllCapabilities {
		level := strings.Count(c.Name, ":")
		cs, ok := capabilitiesByLevel[level]
		if !ok {
			cs = []webCapability{}
		}
		capabilitiesByLevel[level] = append(cs, webCapability{Capability: c})
		if level > maxLevel {
			maxLevel = level
		}
	}
	for level := 0; level <= maxLevel; level++ {
		unhandledReadOnlyNames := make(map[string]*webCapability)
		for _, c := range capabilitiesByLevel[level] {
			if !strings.HasPrefix(c.Name, api.CapabilityReadOnlyPrefix) {
				wc := &WebCapability{ReadWriteCapability: c}
				roc, readOnlyPossible := unhandledReadOnlyNames[c.Name]
				if readOnlyPossible {
					delete(unhandledReadOnlyNames, c.Name)
					wc.ReadOnlyCapability = roc
				}
				parent := searchCapability(c.Name, true)
				if parent == nil {
					allWebCapabilities = append(allWebCapabilities, wc)
				} else if parent.Children == nil {
					parent.Children = []*WebCapability{wc}
				} else {
					parent.Children = append(parent.Children, wc)
				}
				continue
			}
			// readOnly
			var readOnlyCapability webCapability
			c.IsReadOnly = true
			if err := copier.Copy(&readOnlyCapability, c); err != nil {
				panic(err)
			}
			wc := searchCapability(c.Name, false)
			if wc == nil { // capability not added yet
				unhandledReadOnlyNames[c.Name] = &readOnlyCapability
			} else {
				wc.ReadOnlyCapability = &readOnlyCapability
			}
		}
	}
}

func searchCapability(name string, searchParent bool) *WebCapability {
	return searchCapabilityS(allWebCapabilities, name, searchParent)
}
func searchCapabilityS(slice []*WebCapability, name string, searchParent bool) *WebCapability {
	name = strings.TrimPrefix(name, api.CapabilityReadOnlyPrefix)
	for _, c := range slice {
		if !searchParent && c.ReadWriteCapability.Name == name {
			return c
		}
		if strings.HasPrefix(name, c.ReadWriteCapability.Name+":") {
			if searchParent && strings.Count(name, ":")-1 == strings.Count(c.ReadWriteCapability.Name, ":") {
				return c
			}
			return searchCapabilityS(c.Children, name, searchParent)
		}
	}
	return nil
}

func webCapabilityFromCapability(capability api.Capability) *WebCapability {
	return searchCapability(capability.Name, false)
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
	api.CapabilityTokeninfo.Name,
	api.CapabilityTokeninfoIntrospect.Name,
	api.CapabilityTokeninfoHistory.Name,
	api.CapabilityTokeninfoSubtokens.Name,
	api.CapabilityGrantsRead.Name,
	api.CapabilitySSHGrantRead.Name,
}
var warningCapabilities = []string{
	api.CapabilityListMT.Name,
	api.CapabilitySettingsRead.Name,
	api.CapabilitySSHGrant.Name,
}
var dangerCapabilities = []string{
	api.CapabilitySettings.Name,
	api.CapabilityGrants.Name,
}

func (c webCapability) getIntClass() int {
	if c.intClass != nil {
		return *c.intClass
	}
	name := c.Name
	if utils.StringInSlice(name, normalCapabilities) {
		c.intClass = utils.NewInt(intClassNormal)
	}
	if utils.StringInSlice(name, warningCapabilities) {
		c.intClass = utils.NewInt(intClassWarning)
	}
	if utils.StringInSlice(name, dangerCapabilities) {
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

// ColorClass returns the html class for coloring this Capability
func (c webCapability) ColorClass() string {
	return textColorByDanger(c.getDangerLevel())
}

// CapabilityLevel returns a string describing the power of this capability
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

// IsCreateMT checks if this WebCapability is api.CapabilityCreateMT
func (c WebCapability) IsCreateMT() bool {
	return c.ReadWriteCapability.Name == api.CapabilityCreateMT.Name
}
