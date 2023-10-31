package routes

import (
	"fmt"
	"net/url"

	"github.com/oidc-mytoken/utils/utils"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/endpoints/actions/pkg"
	"github.com/oidc-mytoken/server/internal/server/paths"
)

// EndpointURIs
var (
	RedirectURI              string
	ConsentEndpoint          string
	CalendarDownloadEndpoint string
	ActionsEndpoint          string
)

// Init initializes the authcode component
func Init() {
	generalPaths := paths.GetGeneralPaths()
	RedirectURI = utils.CombineURLPath(config.Get().IssuerURL, generalPaths.OIDCRedirectEndpoint)
	ConsentEndpoint = utils.CombineURLPath(config.Get().IssuerURL, generalPaths.ConsentEndpoint)
	CalendarDownloadEndpoint = utils.CombineURLPath(config.Get().IssuerURL, generalPaths.CalendarEndpoint)
	ActionsEndpoint = utils.CombineURLPath(config.Get().IssuerURL, generalPaths.ActionsEndpoint)
}

// ActionsURL builds an action url from a pkg.ActionInfo
func ActionsURL(actionCode pkg.ActionInfo) string {
	params := url.Values{}
	params.Set("action", url.QueryEscape(actionCode.Action))
	params.Set("code", url.QueryEscape(actionCode.Code))
	p := params.Encode()
	return fmt.Sprintf("%s?%s", ActionsEndpoint, p)
}
