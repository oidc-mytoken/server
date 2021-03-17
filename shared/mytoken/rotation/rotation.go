package rotation

type Rotation struct {
	OnAT     bool   `json:"on_AT,omitempty"`
	OnOther  bool   `json:"on_other,omitempty"`
	Lifetime uint64 `json:"lifetime,omitempty"`
}
