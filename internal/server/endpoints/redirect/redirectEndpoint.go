package redirect

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/zachmann/mytoken/internal/server/db"
	"github.com/zachmann/mytoken/internal/server/db/dbrepo/authcodeinforepo"
	"github.com/zachmann/mytoken/internal/server/db/dbrepo/authcodeinforepo/state"
	"github.com/zachmann/mytoken/internal/server/db/dbrepo/supertokenrepo/transfercoderepo"
	"github.com/zachmann/mytoken/internal/server/model"
	"github.com/zachmann/mytoken/internal/server/oidc/authcode"
	"github.com/zachmann/mytoken/internal/server/utils/ctxUtils"
	pkgModel "github.com/zachmann/mytoken/pkg/model"
)

// HandleOIDCRedirect handles redirects from the openid provider after an auth code flow
func HandleOIDCRedirect(ctx *fiber.Ctx) error {
	log.Debug("Handle redirect")
	oidcError := ctx.Query("error")
	state := state.NewState(ctx.Query("state"))
	if len(oidcError) > 0 {
		if len(state.State()) > 0 {
			if err := db.Transact(func(tx *sqlx.Tx) error {
				if err := transfercoderepo.DeleteTransferCodeByState(tx, state); err != nil {
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
			Response: pkgModel.OIDCError(oidcError, oidcErrorDescription),
		}
		return errorRes.Send(ctx)
	}
	code := ctx.Query("code")
	res := authcode.CodeExchange(state, code, *ctxUtils.ClientMetaData(ctx))
	return res.Send(ctx)
}
