package api

// Mytoken is a mytoken Mytoken
type Mytoken struct {
	Issuer               string       `json:"iss"`
	Subject              string       `json:"sub"`
	ExpiresAt            int64        `json:"exp,omitempty"`
	NotBefore            int64        `json:"nbf"`
	IssuedAt             int64        `json:"iat"`
	ID                   string       `json:"jti"`
	SeqNo                uint64       `json:"seq_no"`
	Audience             string       `json:"aud"`
	OIDCSubject          string       `json:"oidc_sub"`
	OIDCIssuer           string       `json:"oidc_iss"`
	Restrictions         Restrictions `json:"restrictions,omitempty"`
	Capabilities         Capabilities `json:"capabilities"`
	SubtokenCapabilities Capabilities `json:"subtoken_capabilities,omitempty"`
	Rotation             Rotation     `json:"rotation,omitempty"`
}

// UsedMytoken is a type for a Mytoken that has been used, it additionally has information how often it has been used
type UsedMytoken struct {
	Mytoken      `json:",inline"`
	Restrictions []UsedRestriction `json:"restrictions,omitempty"`
}

type Rotation struct {
	OnAT     bool   `json:"on_AT,omitempty"`
	OnOther  bool   `json:"on_other,omitempty"`
	Lifetime uint64 `json:"lifetime,omitempty"`
}
