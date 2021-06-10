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
