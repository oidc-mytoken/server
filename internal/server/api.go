package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/utils/utils"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/endpoints/guestmode"
	"github.com/oidc-mytoken/server/internal/endpoints/notification"
	"github.com/oidc-mytoken/server/internal/endpoints/notification/calendar"
	"github.com/oidc-mytoken/server/internal/endpoints/profiles"
	"github.com/oidc-mytoken/server/internal/endpoints/revocation"
	"github.com/oidc-mytoken/server/internal/endpoints/settings"
	"github.com/oidc-mytoken/server/internal/endpoints/settings/email"
	"github.com/oidc-mytoken/server/internal/endpoints/settings/grants"
	"github.com/oidc-mytoken/server/internal/endpoints/settings/grants/ssh"
	"github.com/oidc-mytoken/server/internal/endpoints/token/access"
	"github.com/oidc-mytoken/server/internal/endpoints/token/mytoken"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo"
	"github.com/oidc-mytoken/server/internal/model/version"
	"github.com/oidc-mytoken/server/internal/server/paths"
)

func addAPIRoutes(s fiber.Router) {
	for v := config.Get().API.MinVersion; v <= version.MAJOR; v++ {
		addAPIvXRoutes(s, v)
	}
	guestmode.Init(s)
}

func addAPIvXRoutes(s fiber.Router, version int) {
	apiPaths := paths.GetAPIPaths(version)
	s.Post(apiPaths.MytokenEndpoint, toFiberHandler(mytoken.HandleMytokenEndpoint))
	s.Post(apiPaths.AccessTokenEndpoint, toFiberHandler(access.HandleAccessTokenEndpoint))
	if config.Get().Features.TokenRevocation.Enabled {
		s.Post(apiPaths.RevocationEndpoint, toFiberHandler(revocation.HandleRevoke))
	}
	if config.Get().Features.TransferCodes.Enabled {
		s.Post(apiPaths.TokenTransferEndpoint, toFiberHandler(mytoken.HandleCreateTransferCodeForExistingMytoken))
	}
	if config.Get().Features.TokenInfo.Enabled {
		s.Post(apiPaths.TokenInfoEndpoint, toFiberHandler(tokeninfo.HandleTokenInfo))
	}
	s.Get(apiPaths.UserSettingEndpoint, toFiberHandler(settings.HandleSettings))
	grantPath := utils.CombineURLPath(apiPaths.UserSettingEndpoint, "grants")
	s.Get(grantPath, toFiberHandler(grants.HandleListGrants))
	s.Post(grantPath, toFiberHandler(grants.HandleEnableGrant))
	s.Delete(grantPath, toFiberHandler(grants.HandleDisableGrant))
	if config.Get().Features.SSH.Enabled {
		sshGrantPath := utils.CombineURLPath(grantPath, "ssh")
		s.Get(sshGrantPath, toFiberHandler(ssh.HandleGetSSHInfo))
		s.Post(sshGrantPath, toFiberHandler(ssh.HandlePost))
		s.Delete(sshGrantPath, toFiberHandler(ssh.HandleDeleteSSHKey))
	}
	addProfileEndpointRoutes(s, apiPaths)
	if config.Get().Features.Notifications.AnyEnabled {
		if config.Get().Features.Notifications.ICS.Enabled {
			s.Get(apiPaths.CalendarEndpoint, toFiberHandler(calendar.HandleList))
			s.Post(apiPaths.CalendarEndpoint, toFiberHandler(calendar.HandleAdd))
			s.Get(utils.CombineURLPath(apiPaths.CalendarEndpoint, ":name"), calendar.HandleGet)
			s.Post(utils.CombineURLPath(apiPaths.CalendarEndpoint, ":name"), toFiberHandler(calendar.HandleAddMytoken))
			s.Delete(utils.CombineURLPath(apiPaths.CalendarEndpoint, ":name"), toFiberHandler(calendar.HandleDelete))
		}
		s.Post(apiPaths.NotificationEndpoint, toFiberHandler(notification.HandlePost))
		s.Get(apiPaths.NotificationEndpoint, toFiberHandler(notification.HandleGet))
		s.Get(
			utils.CombineURLPath(apiPaths.NotificationEndpoint, ":code"),
			toFiberHandler(notification.HandleGetByManagementCode),
		)
		s.Delete(
			utils.CombineURLPath(apiPaths.NotificationEndpoint, ":code"),
			toFiberHandler(notification.HandleDeleteByManagementCode),
		)
		s.Post(
			utils.CombineURLPath(apiPaths.NotificationEndpoint, ":code", "nc"),
			toFiberHandler(notification.HandleNotificationUpdateClasses),
		)
		s.Put(
			utils.CombineURLPath(apiPaths.NotificationEndpoint, ":code", "nc"),
			toFiberHandler(notification.HandleNotificationUpdateClasses),
		)
		s.Post(
			utils.CombineURLPath(apiPaths.NotificationEndpoint, ":code", "token"),
			toFiberHandler(notification.HandleNotificationAddToken),
		)
		s.Delete(
			utils.CombineURLPath(apiPaths.NotificationEndpoint, ":code", "token"),
			toFiberHandler(notification.HandleNotificationRemoveToken),
		)
		if config.Get().Features.Notifications.Mail.Enabled {
			s.Get(utils.CombineURLPath(apiPaths.UserSettingEndpoint, "email"), toFiberHandler(email.HandleGet))
			s.Put(utils.CombineURLPath(apiPaths.UserSettingEndpoint, "email"), toFiberHandler(email.HandlePut))
		}
	}
}

func addProfileEndpointRoutes(r fiber.Router, apiPaths paths.APIPaths) {
	if !config.Get().Features.ServerProfiles.Enabled {
		return
	}
	r.Get(apiPaths.ProfilesEndpoint, profiles.HandleGetGroups)
	addProfileGetRoute(r, apiPaths, "profiles", profiles.HandleGetProfiles)
	addProfileGetRoute(r, apiPaths, "capabilities", profiles.HandleGetCapabilities)
	addProfileGetRoute(r, apiPaths, "restrictions", profiles.HandleGetRestrictions)
	addProfileGetRoute(r, apiPaths, "rotation", profiles.HandleGetRotations)
	addProfileAddRoute(r, apiPaths, "profiles", profiles.HandleAddProfile)
	addProfileAddRoute(r, apiPaths, "capabilities", profiles.HandleAddCapabilities)
	addProfileAddRoute(r, apiPaths, "restrictions", profiles.HandleAddRestrictions)
	addProfileAddRoute(r, apiPaths, "rotation", profiles.HandleAddRotation)
	addProfileUpdateRoute(r, apiPaths, "profiles", profiles.HandleUpdateProfile)
	addProfileUpdateRoute(r, apiPaths, "capabilities", profiles.HandleUpdateCapabilities)
	addProfileUpdateRoute(r, apiPaths, "restrictions", profiles.HandleUpdateRestrictions)
	addProfileUpdateRoute(r, apiPaths, "rotation", profiles.HandleUpdateRotation)
	addProfileDeleteRoute(r, apiPaths, "profiles", profiles.HandleDeleteProfile)
	addProfileDeleteRoute(r, apiPaths, "capabilities", profiles.HandleDeleteCapabilities)
	addProfileDeleteRoute(r, apiPaths, "restrictions", profiles.HandleDeleteRestrictions)
	addProfileDeleteRoute(r, apiPaths, "rotation", profiles.HandleDeleteRotation)
}

func addProfileGetRoute(r fiber.Router, apiPaths paths.APIPaths, profileTypePath string, handler fiber.Handler) {
	r.Get(utils.CombineURLPath(apiPaths.ProfilesEndpoint, profileTypePath), handler)
	r.Get(utils.CombineURLPath(apiPaths.ProfilesEndpoint, ":group", profileTypePath), handler)
}

func addProfileDeleteRoute(r fiber.Router, apiPaths paths.APIPaths, profileTypePath string, handler fiber.Handler) {
	r.Delete(
		utils.CombineURLPath(apiPaths.ProfilesEndpoint, profileTypePath, ":id?"),
		returnGroupBasicMiddleware(), userIsGroupMiddleware, handler,
	)
	r.Delete(
		utils.CombineURLPath(apiPaths.ProfilesEndpoint, ":group", profileTypePath, ":id?"),
		returnGroupBasicMiddleware(), userIsGroupMiddleware, handler,
	)
}

func addProfileAddRoute(r fiber.Router, apiPaths paths.APIPaths, profileTypePath string, handler fiber.Handler) {
	r.Post(
		utils.CombineURLPath(apiPaths.ProfilesEndpoint, profileTypePath),
		returnGroupBasicMiddleware(), userIsGroupMiddleware, handler,
	)
	r.Post(
		utils.CombineURLPath(apiPaths.ProfilesEndpoint, ":group", profileTypePath),
		returnGroupBasicMiddleware(), userIsGroupMiddleware, handler,
	)
}

func addProfileUpdateRoute(r fiber.Router, apiPaths paths.APIPaths, profileTypePath string, handler fiber.Handler) {
	r.Put(
		utils.CombineURLPath(apiPaths.ProfilesEndpoint, profileTypePath, ":id?"),
		returnGroupBasicMiddleware(), userIsGroupMiddleware, handler,
	)
	r.Put(
		utils.CombineURLPath(apiPaths.ProfilesEndpoint, ":group", profileTypePath, ":id?"),
		returnGroupBasicMiddleware(), userIsGroupMiddleware, handler,
	)
}
