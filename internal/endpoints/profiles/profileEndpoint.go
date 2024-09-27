package profiles

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db/profilerepo"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// HandleGetGroups handles requests to get the list of groups
func HandleGetGroups(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	groups, err := profilerepo.GetGroups(rlog, nil)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	return model.Response{
		Status:   fiber.StatusOK,
		Response: groups,
	}.Send(ctx)
}

func handleGetProfiles(
	ctx *fiber.Ctx, dbReader func(log.Ext1FieldLogger, *sqlx.Tx, string) ([]api.Profile, error),
) error {
	rlog := logger.GetRequestLogger(ctx)
	group := ctxutils.Params(ctx, "group", "_")
	data, err := dbReader(rlog, nil, group)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	return model.Response{
		Status:   fiber.StatusOK,
		Response: data,
	}.Send(ctx)
}

func handleUpsertProfiles(
	ctx *fiber.Ctx, dbDo func(log.Ext1FieldLogger, *sqlx.Tx, string, string, json.RawMessage) error,
	returnStatus int,
) error {
	rlog := logger.GetRequestLogger(ctx)
	group := ctxutils.Params(ctx, "group", "_")

	var req api.Profile
	err := errors.WithStack(json.Unmarshal(ctx.Body(), &req))
	if err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	err = dbDo(rlog, nil, group, req.Name, req.Payload)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	return model.Response{
		Status: returnStatus,
	}.Send(ctx)
}

func handleDeleteProfiles(
	ctx *fiber.Ctx, dbDo func(log.Ext1FieldLogger, *sqlx.Tx, string, uuid.UUID) error, returnStatus int,
) error {
	rlog := logger.GetRequestLogger(ctx)
	group := ctxutils.Params(ctx, "group", "_")
	id, err := ctxutils.GetID(ctx)
	if err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	err = dbDo(rlog, nil, group, id)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	return model.Response{
		Status: returnStatus,
	}.Send(ctx)
}

// HandleGetProfiles handles get requests for the profiles for a group
func HandleGetProfiles(ctx *fiber.Ctx) error {
	return handleGetProfiles(ctx, profilerepo.GetProfiles)
}

// HandleGetCapabilities handles get requests for the capability templates for a group
func HandleGetCapabilities(ctx *fiber.Ctx) error {
	return handleGetProfiles(ctx, profilerepo.GetCapabilitiesTemplates)
}

// HandleGetRotations handles get requests for the rotation templates for a group
func HandleGetRotations(ctx *fiber.Ctx) error {
	return handleGetProfiles(ctx, profilerepo.GetRotationTemplates)
}

// HandleGetRestrictions handles get requests for the restrictions templates for a group
func HandleGetRestrictions(ctx *fiber.Ctx) error {
	return handleGetProfiles(ctx, profilerepo.GetRestrictionsTemplates)
}

// HandleDeleteProfile handles delete requests for a profile for a group
func HandleDeleteProfile(ctx *fiber.Ctx) error {
	return handleDeleteProfiles(ctx, profilerepo.DeleteProfile, fiber.StatusNoContent)
}

// HandleDeleteRestrictions handles delete requests for a restrictions template for a group
func HandleDeleteRestrictions(ctx *fiber.Ctx) error {
	return handleDeleteProfiles(ctx, profilerepo.DeleteRestrictions, fiber.StatusNoContent)
}

// HandleDeleteRotation handles delete requests for a rotation template for a group
func HandleDeleteRotation(ctx *fiber.Ctx) error {
	return handleDeleteProfiles(ctx, profilerepo.DeleteRotation, fiber.StatusNoContent)
}

// HandleDeleteCapabilities handles delete requests for a capabilities template for a group
func HandleDeleteCapabilities(ctx *fiber.Ctx) error {
	return handleDeleteProfiles(ctx, profilerepo.DeleteCapabilities, fiber.StatusNoContent)
}

// HandleAddProfile handles add requests for a profile for a group
func HandleAddProfile(ctx *fiber.Ctx) error {
	return handleUpsertProfiles(ctx, profilerepo.AddProfile, fiber.StatusCreated)
}

// HandleAddCapabilities handles add requests for a capabilities template for a group
func HandleAddCapabilities(ctx *fiber.Ctx) error {
	return handleUpsertProfiles(ctx, profilerepo.AddCapabilities, fiber.StatusCreated)
}

// HandleAddRestrictions handles add requests for a restrictions template for a group
func HandleAddRestrictions(ctx *fiber.Ctx) error {
	return handleUpsertProfiles(ctx, profilerepo.AddRestrictions, fiber.StatusCreated)
}

// HandleAddRotation handles add requests for a rotation template for a group
func HandleAddRotation(ctx *fiber.Ctx) error {
	return handleUpsertProfiles(ctx, profilerepo.AddRotation, fiber.StatusCreated)
}

// HandleUpdateProfile handles update requests for a profile for a group
func HandleUpdateProfile(ctx *fiber.Ctx) error {
	return handleUpsertProfiles(
		ctx,
		func(rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string, name string, payload json.RawMessage) error {
			id, err := ctxutils.GetID(ctx)
			if err != nil {
				return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
			}
			return profilerepo.UpdateProfile(rlog, tx, group, id, name, payload)
		}, fiber.StatusOK,
	)
}

// HandleUpdateCapabilities handles update requests for a capabilities template for a group
func HandleUpdateCapabilities(ctx *fiber.Ctx) error {
	return handleUpsertProfiles(
		ctx,
		func(rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string, name string, payload json.RawMessage) error {
			id, err := ctxutils.GetID(ctx)
			if err != nil {
				return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
			}
			return profilerepo.UpdateCapabilities(rlog, tx, group, id, name, payload)
		}, fiber.StatusOK,
	)
}

// HandleUpdateRestrictions handles update requests for a restrictions template for a group
func HandleUpdateRestrictions(ctx *fiber.Ctx) error {
	return handleUpsertProfiles(
		ctx,
		func(rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string, name string, payload json.RawMessage) error {
			id, err := ctxutils.GetID(ctx)
			if err != nil {
				return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
			}
			return profilerepo.UpdateRestrictions(rlog, tx, group, id, name, payload)
		}, fiber.StatusOK,
	)
}

// HandleUpdateRotation handles update requests for a rotation template for a group
func HandleUpdateRotation(ctx *fiber.Ctx) error {
	return handleUpsertProfiles(
		ctx,
		func(rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string, name string, payload json.RawMessage) error {
			id, err := ctxutils.GetID(ctx)
			if err != nil {
				return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
			}
			return profilerepo.UpdateRotation(rlog, tx, group, id, name, payload)
		}, fiber.StatusOK,
	)
}
