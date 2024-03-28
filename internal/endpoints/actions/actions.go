package actions

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/actionrepo"
	"github.com/oidc-mytoken/server/internal/endpoints/actions/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/server/routes"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// HandleActions is the main entry function to handle the different actions of the action endpoint
func HandleActions(ctx *fiber.Ctx) error {
	actionInfo := pkg.CtxGetActionInfo(ctx)
	switch actionInfo.Action {
	case pkg.ActionRecreate:
		return handleRecreate(ctx, actionInfo.Code)
	case pkg.ActionVerifyEmail:
		return handleVerifyEmail(ctx, actionInfo.Code)
	case pkg.ActionRemoveFromCalendar:
		return handleRemoveFromCalendar(ctx, actionInfo.Code)
	}
	return ctxutils.RenderErrorPage(
		ctx, fiber.StatusBadRequest, model.BadRequestError("unknown action").
			CombinedMessage(),
	)
}

func handleRecreate(ctx *fiber.Ctx, code string) error {
	return ctxutils.RenderErrorPage(ctx, fiber.StatusNotImplemented, api.ErrorNYI.CombinedMessage())
}
func handleVerifyEmail(ctx *fiber.Ctx, code string) error {
	rlog := logger.GetRequestLogger(ctx)
	verified, err := actionrepo.VerifyMail(rlog, nil, code)
	if err != nil {
		return ctxutils.RenderInternalServerErrorPage(ctx, err)
	}
	if !verified {
		return ctxutils.RenderErrorPage(ctx, http.StatusBadRequest, "code not valid or expired")
	}
	return ctxutils.RenderErrorPage(
		ctx, http.StatusOK, "The email address was successfully verified.", "Email Verified",
	)
}

func handleRemoveFromCalendar(ctx *fiber.Ctx, code string) error {
	rlog := logger.GetRequestLogger(ctx)
	err := actionrepo.UseRemoveCalendarCode(rlog, nil, code)
	if err != nil {
		return ctxutils.RenderInternalServerErrorPage(ctx, err)
	}
	return ctxutils.RenderErrorPage(
		ctx, http.StatusOK, "The token was successfully removed from the calendar.", "Token Removed from Calendar",
	)
}

// CreateVerifyEmail creates an action url for verifying a mail address
func CreateVerifyEmail(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID) (string, error) {
	code := pkg.ActionInfo{
		Action: pkg.ActionVerifyEmail,
		Code:   pkg.NewCode(),
	}
	if err := actionrepo.AddVerifyEmailCode(rlog, tx, mtID, code.Code, pkg.CodeLifetimes[code.Action]); err != nil {
		return "", err
	}
	return routes.ActionsURL(code), nil
}

// CreateRecreateToken creates an action url for recreating a mytoken
func CreateRecreateToken(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID) (string, error) {
	code := pkg.ActionInfo{
		Action: pkg.ActionRecreate,
		Code:   pkg.NewCode(),
	}
	if err := actionrepo.AddRecreateTokenCode(rlog, tx, mtID, code.Code); err != nil {
		return "", err
	}
	return routes.ActionsURL(code), nil
}

// CreateRemoveFromCalendar creates an action url for removing a token from a calendar
func CreateRemoveFromCalendar(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, calendarName string) (
	string,
	error,
) {
	code := pkg.ActionInfo{
		Action: pkg.ActionRemoveFromCalendar,
		Code:   pkg.NewCode(),
	}
	if err := actionrepo.AddRemoveFromCalendarCode(rlog, tx, mtID, code.Code, calendarName); err != nil {
		return "", err
	}
	return routes.ActionsURL(code), nil
}
