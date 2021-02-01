package server

import (
	"github.com/gofiber/fiber/v2"

	dbhelper "github.com/oidc-mytoken/server/internal/db/dbrepo/supertokenrepo/supertokenrepohelper"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/utils/ctxUtils"
	supertoken "github.com/oidc-mytoken/server/shared/supertoken/pkg"
)

func handleIndex(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		"logged-in": false,
	}
	return ctx.Render("sites/index", binding, "layouts/main")
}

func handleHome(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		"logged-in": true,
	}
	return ctx.Render("sites/home", binding, "layouts/main")
}

func handleNativeCallback(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		"empty-navbar": true, //TODO implement
	}
	return ctx.Render("sites/native", binding, "layouts/main")
}

func handleTestTokenInfo(ctx *fiber.Ctx) error {
	tok := ctxUtils.GetSuperToken(ctx)
	if tok == nil {
		return model.Response{
			Status: fiber.StatusUnauthorized,
		}.Send(ctx)
	}

	st, err := supertoken.ParseJWT(string(*tok))
	if err != nil {
		return model.Response{
			Status: fiber.StatusUnauthorized,
		}.Send(ctx)
	}

	revoked, dbErr := dbhelper.CheckTokenRevoked(st.ID)
	if dbErr != nil {
		return model.ErrorToInternalServerErrorResponse(dbErr).Send(ctx)
	}
	if revoked {
		return model.Response{
			Status: fiber.StatusUnauthorized,
		}.Send(ctx)
	}
	return model.Response{
		Status: fiber.StatusNoContent,
	}.Send(ctx)
}
