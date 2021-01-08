package pkg

import (
	"github.com/zachmann/mytoken/internal/model"
)

// CreateTransferCodeRequest is a request to create a new transfer code from an existing super token
type CreateTransferCodeRequest struct {
	SuperToken string `json:"super_token"` // we use string and not token.Token because the token can also be in the Auth Header and there it is a string
}

// ExchangeTransferCodeRequest is a request to exchange a transfer code for the super token
type ExchangeTransferCodeRequest struct {
	GrantType    model.GrantType `json:"grant_type"`
	TransferCode string          `json:"transfer_code"`
}
