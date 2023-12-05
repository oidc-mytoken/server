package routes

import (
	"github.com/oidc-mytoken/utils/utils"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/server/paths"
)

var RedirectURI string
var ConsentEndpoint string

// Init initializes the authcode component
func Init() {
	generalPaths := paths.GetGeneralPaths()
	RedirectURI = utils.CombineURLPath(config.Get().IssuerURL, generalPaths.OIDCRedirectEndpoint)
	ConsentEndpoint = utils.CombineURLPath(config.Get().IssuerURL, generalPaths.ConsentEndpoint)
}
