package capabilities

// Capabilities is a slice of Capability
type Capabilities []Capability

// Capability is a capability string
type Capability string

// Tighten thightens two set of Capabilities into one new
func Tighten(a, b Capabilities) (res Capabilities) {
	for _, bb := range b {
		for _, aa := range a {
			if bb == aa {
				res = append(res, bb)
			}
		}
	}
	return
}
