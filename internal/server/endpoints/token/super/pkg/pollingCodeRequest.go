package pkg

import (
	"github.com/zachmann/mytoken/internal/model"
)

// PollingCodeRequest is a polling code request
type PollingCodeRequest struct {
	GrantType   model.GrantType `json:"grant_type"`
	PollingCode string          `json:"polling_code"`
}
