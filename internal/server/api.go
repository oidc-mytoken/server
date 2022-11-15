package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/utils/utils"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/endpoints/revocation"
	"github.com/oidc-mytoken/server/internal/endpoints/settings"
	"github.com/oidc-mytoken/server/internal/endpoints/settings/grants"
	"github.com/oidc-mytoken/server/internal/endpoints/settings/grants/ssh"
	"github.com/oidc-mytoken/server/internal/endpoints/token/access"
	"github.com/oidc-mytoken/server/internal/endpoints/token/mytoken"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo"
	"github.com/oidc-mytoken/server/internal/model/version"
	"github.com/oidc-mytoken/server/internal/server/routes"
)

func addAPIRoutes(s fiber.Router) {
	for v := config.Get().API.MinVersion; v <= version.MAJOR; v++ {
		addAPIvXRoutes(s, v)
	}
}

func addAPIvXRoutes(s fiber.Router, version int) {
	apiPaths := routes.GetAPIPaths(version)
	s.Post(apiPaths.MytokenEndpoint, mytoken.HandleMytokenEndpoint)
	s.Post(apiPaths.AccessTokenEndpoint, access.HandleAccessTokenEndpoint)
	if config.Get().Features.TokenRevocation.Enabled {
		s.Post(apiPaths.RevocationEndpoint, revocation.HandleRevoke)
	}
	if config.Get().Features.TransferCodes.Enabled {
		s.Post(apiPaths.TokenTransferEndpoint, mytoken.HandleCreateTransferCodeForExistingMytoken)
	}
	if config.Get().Features.TokenInfo.Enabled {
		s.Post(apiPaths.TokenInfoEndpoint, tokeninfo.HandleTokenInfo)
	}
	s.Get(apiPaths.UserSettingEndpoint, settings.HandleSettings)
	grantPath := utils.CombineURLPath(apiPaths.UserSettingEndpoint, "grants")
	s.Get(grantPath, grants.HandleListGrants)
	s.Post(grantPath, grants.HandleEnableGrant)
	s.Delete(grantPath, grants.HandleDisableGrant)
	if config.Get().Features.SSH.Enabled {
		sshGrantPath := utils.CombineURLPath(grantPath, "ssh")
		s.Get(sshGrantPath, ssh.HandleGetSSHInfo)
		s.Post(sshGrantPath, ssh.HandlePost)
		s.Delete(sshGrantPath, ssh.HandleDeleteSSHKey)
	}
}
