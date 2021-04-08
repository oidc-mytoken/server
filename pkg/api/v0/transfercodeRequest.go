package api

// CreateTransferCodeRequest is a request to create a new transfer code from an existing mytoken
type CreateTransferCodeRequest struct {
	Mytoken string `json:"mytoken"` // we use string and not token.Token because the token can also be in the Auth Header and there it is a string
}

// ExchangeTransferCodeRequest is a request to exchange a transfer code for the mytoken
type ExchangeTransferCodeRequest struct {
	GrantType    string `json:"grant_type"`
	TransferCode string `json:"transfer_code"`
}
