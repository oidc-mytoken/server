package pkg

type AuthCodeFlowResponse struct {
	AuthorizationURL     string `json:"authorization_url"`
	PollingCode          string `json:"polling_code,omitempty"`
	PollingCodeExpiresIn int64  `json:"polling_code_expires_in,omitempty"`
	PollingInterval      int64  `json:"polling_interval,omitempty"`
}
