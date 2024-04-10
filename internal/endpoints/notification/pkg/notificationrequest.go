package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
)

// SubscribeNotificationRequest is type holding the request to create different notifications
type SubscribeNotificationRequest struct {
	api.SubscribeNotificationRequest
	Mytoken universalmytoken.UniversalMytoken `json:"mytoken" xml:"mytoken" form:"mytoken"`
	MomID   mtid.MOMID                        `json:"mom_id" xml:"mom_id" form:"mom_id"`
}

// NotificationsListResponse is a type holding the response to a notification list request
type NotificationsListResponse struct {
	api.NotificationsListResponse
	pkg.OnlyTokenUpdateRes
}

// NotificationAddTokenRequest is a request object for adding a mytoken to an existing notification
type NotificationAddTokenRequest struct {
	api.NotificationAddTokenRequest
	Mytoken universalmytoken.UniversalMytoken `json:"mytoken" xml:"mytoken" form:"mytoken"`
	MomID   mtid.MOMID                        `json:"mom_id" xml:"mom_id" form:"mom_id"`
}

// NotificationRemoveTokenRequest is a request object for removing a mytoken from a notification
type NotificationRemoveTokenRequest struct {
	api.NotificationRemoveTokenRequest
	Mytoken universalmytoken.UniversalMytoken `json:"mytoken" xml:"mytoken" form:"mytoken"`
	MomID   mtid.MOMID                        `json:"mom_id" xml:"mom_id" form:"mom_id"`
}

// NotificationsCreateResponse is a type holding the response to a notification creation request
type NotificationsCreateResponse struct {
	api.NotificationsCreateResponse
	pkg.OnlyTokenUpdateRes
}
