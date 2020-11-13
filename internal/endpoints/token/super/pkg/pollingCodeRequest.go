package pkg

import (
	"github.com/zachmann/mytoken/internal/model"
)

type PollingCodeRequest struct {
	GrantType   model.GrantType `json:"grant_type"`
	PollingCode string          `json:"polling_code"`
}
