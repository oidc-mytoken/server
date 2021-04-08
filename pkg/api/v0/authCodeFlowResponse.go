package api

// AuthCodeFlowResponse is the response to an authorization code flow request
type AuthCodeFlowResponse struct {
	AuthorizationURL string `json:"authorization_url"`
	PollingInfo
}

// PollingInfo holds all response information about polling codes
type PollingInfo struct {
	PollingCode          string `json:"polling_code,omitempty"`
	PollingCodeExpiresIn int64  `json:"polling_code_expires_in,omitempty"`
	PollingInterval      int64  `json:"polling_interval,omitempty"`
}
