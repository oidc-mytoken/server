package redirect

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/zachmann/mytoken/internal/db"
	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/oidc/authcode"
	"github.com/zachmann/mytoken/internal/utils/ctxUtils"
)

// HandleOIDCRedirect handles redirects from the openid provider after an auth code flow
func HandleOIDCRedirect(ctx *fiber.Ctx) error {
	log.Debug("Handle redirect")
	oidcError := ctx.Query("error")
	state := ctx.Query("state")
	if oidcError != "" {
		if state != "" {
			_ = db.Transact(func(tx *sqlx.Tx) error {
				if _, err := tx.Exec(`DELETE FROM PollingCodes WHERE id=(SELECT polling_code_id FROM AuthInfo WHERE state=?)`, state); err != nil {
					return err
				}
				if _, err := tx.Exec(`DELETE FROM AuthInfo WHERE state=?`, state); err != nil {
					return err
				}
				return nil
			})
		}
		oidcErrorDescription := ctx.Query("error_description")
		errorRes := model.Response{
			Status:   fiber.StatusInternalServerError,
			Response: model.OIDCError(oidcError, oidcErrorDescription),
		}
		return errorRes.Send(ctx)
	}
	code := ctx.Query("code")
	res := authcode.CodeExchange(state, code, *ctxUtils.ClientMetaData(ctx))
	return res.Send(ctx)
}
