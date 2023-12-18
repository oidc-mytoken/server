package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
)

// SubscribeNotificationRequest is type holding the request to create different notifications
type SubscribeNotificationRequest struct {
	api.SubscribeNotificationRequest
	Mytoken universalmytoken.UniversalMytoken `json:"mytoken" xml:"mytoken" form:"mytoken"`
	MomID   mtid.MOMID                        `json:"mom_id" xml:"mom_id" form:"mom_id"`
}
