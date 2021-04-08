package api

// TransferCodeResponse is the response to a transfer code request
type TransferCodeResponse struct {
	MytokenType  string `json:"mytoken_type"`
	TransferCode string `json:"transfer_code"`
	ExpiresIn    uint64 `json:"expires_in"`
}
