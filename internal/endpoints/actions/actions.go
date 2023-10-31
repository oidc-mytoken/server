package actions

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/actionrepo"
	"github.com/oidc-mytoken/server/internal/endpoints/actions/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/server/routes"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// HandleActions is the main entry function to handle the different actions of the action endpoint
func HandleActions(ctx *fiber.Ctx) error {
	actionInfo := pkg.CtxGetActionInfo(ctx)
	rlog := logger.GetRequestLogger(ctx)
	switch actionInfo.Action {
	case pkg.ActionRecreate:
		return handleRecreate(rlog, actionInfo.Code).Send(ctx)
	case pkg.ActionUnsubscribe:
		return handleUnsubscribe(rlog, actionInfo.Code).Send(ctx)
	case pkg.ActionVerifyEmail:
		return handleVerifyEmail(rlog, actionInfo.Code).Send(ctx)
	case pkg.ActionRemoveFromCalendar:
		return handleRemoveFromCalendar(rlog, actionInfo.Code).Send(ctx)
	}
	return model.Response{
		Status:   http.StatusBadRequest,
		Response: model.BadRequestError("unknown action"),
	}.Send(ctx)
}

func handleRecreate(rlog log.Ext1FieldLogger, code string) model.Response {
	return model.ResponseNYI
}
func handleVerifyEmail(rlog log.Ext1FieldLogger, code string) *model.Response {
	verified, err := actionrepo.VerifyMail(rlog, nil, code)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if !verified {
		return &model.Response{
			Status:   http.StatusBadRequest,
			Response: model.BadRequestError("code not valid or expired"),
		}
	}
	return &model.Response{
		Status: http.StatusOK,
	}
}
func handleUnsubscribe(rlog log.Ext1FieldLogger, code string) model.Response {
	return model.ResponseNYI
}
func handleRemoveFromCalendar(rlog log.Ext1FieldLogger, code string) model.Response {
	return model.ResponseNYI
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
