package redirect

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/valyala/fasthttp"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo/state"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	pkgModel "github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/authcode"
	"github.com/oidc-mytoken/server/internal/server/httpstatus"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// HandleOIDCRedirect handles redirects from the openid provider after an auth code flow
func HandleOIDCRedirect(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle redirect")
	oidcError := ctx.Query("error")
	oState := state.NewState(ctx.Query("state"))
	if oidcError != "" {
		if oState.State() != "" {
			if err := db.Transact(
				rlog, func(tx *sqlx.Tx) error {
					if err := transfercoderepo.DeleteTransferCodeByState(rlog, tx, oState); err != nil {
						return err
					}
					return authcodeinforepo.DeleteAuthFlowInfoByState(rlog, tx, oState)
				},
			); err != nil {
				rlog.Errorf("%s", errorfmt.Full(err))
			}
		}
		oidcErrorDescription := ctx.Query("error_description")
		return ctx.Status(httpstatus.StatusOIDPError).Render(
			"sites/error", map[string]interface{}{
				"empty-navbar":  true,
				"error-heading": "OIDC error",
				"msg":           pkgModel.OIDCError(oidcError, oidcErrorDescription).CombinedMessage(),
			}, "layouts/main",
		)
	}
	code := ctx.Query("code")
	res := authcode.CodeExchange(rlog, oState, code, *ctxutils.ClientMetaData(ctx))

	if fasthttp.StatusCodeIsRedirect(res.Status) {
		return res.Send(ctx)
	}
	return ctx.Status(res.Status).Render(
		"sites/error", map[string]interface{}{
			"empty-navbar":  true,
			"error-heading": http.StatusText(res.Status),
			"msg":           res.Response.(api.Error).CombinedMessage(),
		}, "layouts/main",
	)
}
