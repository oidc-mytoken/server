package pkg

import (
	"github.com/zachmann/mytoken/internal/model"
)

// TransferCodeResponse is the response to a transfer code request
type TransferCodeResponse struct {
	SuperTokenType model.ResponseType `json:"super_token_type"`
	TransferCode   string             `json:"transfer_code"`
	ExpiresIn      uint64             `json:"expires_in"`
}
