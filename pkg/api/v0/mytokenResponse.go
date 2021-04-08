package api

// MytokenResponse is a response to a mytoken request
type MytokenResponse struct {
	Mytoken              string       `json:"mytoken,omitempty"`
	MytokenType          string       `json:"mytoken_type"`
	TransferCode         string       `json:"transfer_code,omitempty"`
	ExpiresIn            uint64       `json:"expires_in,omitempty"`
	Restrictions         Restrictions `json:"restrictions,omitempty"`
	Capabilities         Capabilities `json:"capabilities,omitempty"`
	SubtokenCapabilities Capabilities `json:"subtoken_capabilities,omitempty"`
}
