package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
)

type AddMytokenToCalendarRequest struct {
	api.AddMytokenToCalendarRequest
	MomID mtid.MOMID `json:"mom_id" xml:"mom_id" form:"mom_id"`
}
