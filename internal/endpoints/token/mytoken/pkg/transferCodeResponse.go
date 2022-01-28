package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/shared/model"
)

// TransferCodeResponse is the response to a transfer code request
type TransferCodeResponse struct {
	api.TransferCodeResponse `json:",inline"`
	MytokenType              model.ResponseType `json:"mytoken_type"`
}
