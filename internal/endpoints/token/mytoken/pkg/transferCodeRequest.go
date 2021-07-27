package pkg

import (
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/server/shared/model"
)

// ExchangeTransferCodeRequest is a request to exchange a transfer code for the mytoken
type ExchangeTransferCodeRequest struct {
	api.ExchangeTransferCodeRequest `json:",inline"`
	GrantType                       model.GrantType `json:"grant_type"`
}

// NewExchangeTransferCodeRequest returns a new ExchangeTransferCodeRequest
func NewExchangeTransferCodeRequest() ExchangeTransferCodeRequest {
	return ExchangeTransferCodeRequest{
		GrantType: -1, // This value will remain if grant_type is not contained in the request. We have to set it to -1,
		// because the default of 0 would be a valid GrantType
	}
}
