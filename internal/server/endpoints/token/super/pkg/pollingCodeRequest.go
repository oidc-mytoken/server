package pkg

import (
	"github.com/zachmann/mytoken/pkg/model"
)

// PollingCodeRequest is a polling code request
type PollingCodeRequest struct {
	GrantType   model.GrantType `json:"grant_type"`
	PollingCode string          `json:"polling_code"`
}
