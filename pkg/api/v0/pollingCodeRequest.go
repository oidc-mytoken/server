package api

// PollingCodeRequest is a polling code request
type PollingCodeRequest struct {
	GrantType   string `json:"grant_type"`
	PollingCode string `json:"polling_code"`
}
