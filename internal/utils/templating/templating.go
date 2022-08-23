package templating

// Collapsable is a struct for controlling collapsable content in the templates
type Collapsable struct {
	All           bool
	CollapsRot    bool
	CollapseRestr bool
	CollapseCaps  bool
}

// Rotation checks if rotations should be collapsed
func (c Collapsable) Rotation() bool {
	return c.All || c.CollapsRot
}

// Restrictions checks if restrictions should be collapsed
func (c Collapsable) Restrictions() bool {
	return c.All || c.CollapseRestr
}

// Capabilities checks if capabilities should be collapsed
func (c Collapsable) Capabilities() bool {
	return c.All || c.CollapseCaps
}
