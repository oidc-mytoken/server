package api

import (
	"encoding/json"
	"fmt"
)

// Mytoken is a mytoken Mytoken
type Mytoken struct {
	Version              TokenVersion `json:"ver"`
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

var TokenVer = TokenVersion{
	Major: 0,
	Minor: 1,
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

type TokenVersion struct {
	Major int
	Minor int
}

const vFmt = "%d.%d"

func (v TokenVersion) String() string {
	return fmt.Sprintf(vFmt, v.Major, v.Minor)
}
func (v TokenVersion) Version() string {
	return v.String()
}
func (v TokenVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}
func (v *TokenVersion) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	_, err := fmt.Sscanf(str, vFmt, &v.Major, &v.Minor)
	return err
}
