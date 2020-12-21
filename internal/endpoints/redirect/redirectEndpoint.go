package redirect

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/zachmann/mytoken/internal/db"
	"github.com/zachmann/mytoken/internal/db/dbrepo/authcodeinforepo"
	"github.com/zachmann/mytoken/internal/db/dbrepo/pollingcoderepo"
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
			if err := db.Transact(func(tx *sqlx.Tx) error {
				if err := pollingcoderepo.DeletePollingCodeByState(tx, state); err != nil {
					return err
				}
				if err := authcodeinforepo.DeleteAuthFlowInfoByState(tx, state); err != nil {
					return err
				}
				return nil
			}); err != nil {
				log.WithError(err).Error()
			}
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
