package redirect

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/zachmann/mytoken/internal/db"
	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/oidc/authcode"
)

func HandleOIDCRedirect(ctx *fiber.Ctx) error {
	log.Print("Handle redirect")
	oidcError := ctx.Query("error")
	state := ctx.Query("state")
	log.Printf("error: '%s'", oidcError)
	log.Printf("state: '%s'", state)
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
	networkData := model.NetworkData{
		IP:        ctx.IP(),
		UserAgent: string(ctx.Request().Header.UserAgent()),
	}
	res := authcode.CodeExchange(state, code, networkData)
	return res.Send(ctx)
}
