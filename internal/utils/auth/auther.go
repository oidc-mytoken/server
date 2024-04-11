package auth

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/eventrepo"
	dbhelper "github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	notifier "github.com/oidc-mytoken/server/internal/notifier/client"
	provider2 "github.com/oidc-mytoken/server/internal/oidc/provider"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/iputils"
)

// RequireGrantType checks that the passed model.GrantType are the same, and returns an error model.Response if not
func RequireGrantType(rlog log.Ext1FieldLogger, want, got model.GrantType) *model.Response {
	if got != want {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnsupportedGrantType,
		}
	}
	rlog.Trace("Checked grant type")
	return nil
}

// RequireMytoken checks the passed universalmytoken.UniversalMytoken and if needed other request parameters like
// authorization header and cookie value for a mytoken string. The mytoken string is parsed and if not valid an error
// model.Response is returned.
func RequireMytoken(rlog log.Ext1FieldLogger, reqToken *universalmytoken.UniversalMytoken, ctx *fiber.Ctx) (
	*mytoken.Mytoken, *model.Response,
) {
	if reqToken.JWT == "" {
		t, found := ctxutils.GetMytoken(ctx)
		if t == nil {
			errDesc := "no mytoken found in request"
			if found {
				errDesc = "token not valid"
			}
			return nil, &model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: model.InvalidTokenError(errDesc),
			}
		}
		*reqToken = *t
	}

	mt, err := mytoken.ParseJWT(reqToken.JWT)
	if err != nil {
		return nil, &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model.InvalidTokenError(errorfmt.Error(err)),
		}
	}
	rlog.Trace("Parsed mytoken")
	return mt, nil
}

// RequireMytokenNotRevoked checks that the passed mytoken.Mytoken was not revoked, if it was an error model.Response is
// returned.
func RequireMytokenNotRevoked(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mt *mytoken.Mytoken,
	clientData *api.ClientMetaData,
) *model.Response {
	revoked, dbErr := dbhelper.CheckTokenRevoked(rlog, tx, mt.ID, mt.SeqNo, mt.Rotation)
	if dbErr != nil {
		rlog.Errorf("%s", errorfmt.Full(dbErr))
		return model.ErrorToInternalServerErrorResponse(dbErr)
	}
	if revoked {
		_ = notifier.SendNotificationsForSubClass(
			rlog, tx, mt.ID, api.NotificationClassRevokedUsage, clientData,
			nil, nil,
		)
		return &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model.InvalidTokenError(""),
		}
	}
	rlog.Trace("Checked mytoken not revoked")
	checkIPUnusual(rlog, tx, mt.ID, clientData)
	return nil
}

// RequireValidMytoken checks the passed universalmytoken.UniversalMytoken and if needed other request parameters like
// authorization header and cookie value for a mytoken string. The mytoken string is parsed and if not valid an error
// model.Response is returned. RequireValidMytoken also asserts that the mytoken.Mytoken was not revoked.
func RequireValidMytoken(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, reqToken *universalmytoken.UniversalMytoken, ctx *fiber.Ctx,
) (
	*mytoken.Mytoken, *model.Response,
) {
	mt, errRes := RequireMytoken(rlog, reqToken, ctx)
	if errRes != nil {
		return nil, errRes
	}
	return mt, RequireMytokenNotRevoked(rlog, tx, mt, ctxutils.ClientMetaData(ctx))
}

// RequireMatchingIssuer checks that the OIDC issuer from a mytoken is the same as the issuer string in a request (if
// given). RequireMatchingIssuer also checks that the issuer is valid for this mytoken instance.
func RequireMatchingIssuer(rlog log.Ext1FieldLogger, mtOIDCIssuer string, requestIssuer *string) (
	model.Provider, *model.Response,
) {
	if *requestIssuer == "" {
		*requestIssuer = mtOIDCIssuer
		rlog.Trace("Checked issuer (was not given)")
	}
	if *requestIssuer != mtOIDCIssuer {
		return nil, model.BadRequestErrorResponse("token not for specified issuer")
	}
	provider := provider2.GetProvider(*requestIssuer)
	if provider == nil {
		return nil, &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnknownIssuer,
		}
	}
	rlog.Trace("Checked issuer")
	return provider, nil
}

// RequireCapability checks that the passed mytoken.Mytoken has the required api.Capability and returns an error
// model.Response if not
func RequireCapability(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, capability api.Capability, mt *mytoken.Mytoken,
	clientData *api.ClientMetaData,
) *model.Response {
	if !mt.Capabilities.Has(capability) {
		_ = notifier.SendNotificationsForSubClass(
			rlog, tx, mt.ID, api.NotificationClassInsufficientCapabilities, clientData,
			model.KeyValues{
				{
					Key:   "Needed Capability",
					Value: capability.Name,
				},
			}, nil,
		)
		return &model.Response{
			Status:   fiber.StatusForbidden,
			Response: api.ErrorInsufficientCapabilities,
		}
	}
	rlog.Trace("Checked capability")
	return nil
}

func requireUseableRestriction(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mt *mytoken.Mytoken, clientData *api.ClientMetaData, scopes, auds []string,
	at bool,
) (*restrictions.Restriction, *model.Response) {
	if len(mt.Restrictions) == 0 {
		return nil, nil
	}
	getUseableRestrictions := mt.Restrictions.GetValidForOther
	if at {
		getUseableRestrictions = mt.Restrictions.GetValidForAT
	}
	// WithScopes and WithAudience don't tighten the restrictions if nil is passed
	useableRestrictions := getUseableRestrictions(rlog, tx, clientData.IP, mt.ID).WithScopes(
		rlog, scopes,
	).WithAudiences(
		rlog, auds,
	)
	if len(useableRestrictions) == 0 {
		_ = notifier.SendNotificationsForSubClass(
			rlog, tx, mt.ID, api.NotificationClassRestrictedUsages, clientData, nil, nil,
		)
		return nil, &model.Response{
			Status:   fiber.StatusForbidden,
			Response: api.ErrorUsageRestricted,
		}
	}
	rlog.Trace("Checked mytoken restrictions")
	return useableRestrictions[0], nil
}

func checkIPUnusual(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, clientData *api.ClientMetaData,
) {
	rlog.WithField("ip", clientData.IP).Debug("Checking if this is an unusual ip")
	_ = notifier.SendNotificationsForSubClass(
		rlog, tx, mtID, api.NotificationClassUnusualIPs, clientData, nil,
		func() (bool, error) {
			ips, err := eventrepo.GetPreviouslyUsedIPs(rlog, tx, mtID)
			if err != nil {
				return false, err
			}
			if len(ips) == 0 {
				return false, nil
			}
			return !iputils.IPIsIn(clientData.IP, ips), nil
		},
	)
}

// RequireUsableRestriction checks that the mytoken.Mytoken's restrictions allow the usage
func RequireUsableRestriction(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mt *mytoken.Mytoken, clientData *api.ClientMetaData, scopes, auds []string,
	capability api.Capability,
) (*restrictions.Restriction, *model.Response) {
	return requireUseableRestriction(rlog, tx, mt, clientData, scopes, auds, capability == api.CapabilityAT)
}

// RequireUsableRestrictionAT checks that the mytoken.Mytoken's restrictions allow the AT usage
func RequireUsableRestrictionAT(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mt *mytoken.Mytoken, clientData *api.ClientMetaData, scopes, auds []string,
) (*restrictions.Restriction, *model.Response) {
	return requireUseableRestriction(rlog, tx, mt, clientData, scopes, auds, true)
}

// RequireUsableRestrictionOther checks that the mytoken.Mytoken's restrictions allow the non-AT usage
func RequireUsableRestrictionOther(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mt *mytoken.Mytoken, clientData *api.ClientMetaData,
) (*restrictions.Restriction, *model.Response) {
	return requireUseableRestriction(rlog, tx, mt, clientData, nil, nil, false)
}

// RequireCapabilityAndRestriction checks the mytoken.Mytoken's capability and restrictions
func RequireCapabilityAndRestriction(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mt *mytoken.Mytoken, clientData *api.ClientMetaData, scopes, auds []string,
	capability api.Capability,
) (*restrictions.Restriction, *model.Response) {
	if errRes := RequireCapability(rlog, tx, capability, mt, clientData); errRes != nil {
		return nil, errRes
	}
	return RequireUsableRestriction(rlog, tx, mt, clientData, scopes, auds, capability)
}

// RequireCapabilityAndRestrictionOther checks the mytoken.Mytoken's capability and restrictions
func RequireCapabilityAndRestrictionOther(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mt *mytoken.Mytoken, clientData *api.ClientMetaData,
	capability api.Capability,
) (*restrictions.Restriction, *model.Response) {
	if errRes := RequireCapability(rlog, tx, capability, mt, clientData); errRes != nil {
		return nil, errRes
	}
	return RequireUsableRestrictionOther(rlog, tx, mt, clientData)
}

// RequireMytokensForSameUser checks that the two passed mtid.MTID are mytokens for the same user and returns an error
// model.Response if not
func RequireMytokensForSameUser(rlog log.Ext1FieldLogger, tx *sqlx.Tx, id1, id2 mtid.MTID) *model.Response {
	same, err := dbhelper.CheckMytokensAreForSameUser(rlog, tx, id1, id2)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if !same {
		return &model.Response{
			Status: fiber.StatusForbidden,
			Response: api.Error{
				Error:            api.ErrorStrInvalidGrant,
				ErrorDescription: "The provided token cannot be used to manage this mom_id",
			},
		}
	}
	rlog.Trace("Checked mytokens are for same user")
	return nil
}

func RequireMytokenIsParentOrCapability(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, capabilityIfParent,
	capabilityIfNotParent api.Capability,
	mt *mytoken.Mytoken, momID mtid.MTID, clientData *api.ClientMetaData,
) *model.Response {
	isParent, err := dbhelper.MOMIDHasParent(rlog, tx, momID.Hash(), mt.ID)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if isParent && mt.Capabilities.Has(capabilityIfParent) {
		rlog.Trace("Checked mytoken is parent or has capability")
		return nil
	}
	if mt.Capabilities.Has(capabilityIfNotParent) {
		rlog.Trace("Checked mytoken is parent or has capability")
		return nil
	}
	_ = notifier.SendNotificationsForSubClass(
		rlog, tx, mt.ID, api.NotificationClassInsufficientCapabilities, clientData,
		model.KeyValues{
			{
				Key:   "Needed Capability",
				Value: capabilityIfNotParent.Name,
			},
		}, nil,
	)
	return &model.Response{
		Status: fiber.StatusForbidden,
		Response: api.Error{
			Error: api.ErrorStrInsufficientCapabilities,
			ErrorDescription: fmt.Sprintf(
				"The provided token is neither a parent of the subject token"+
					" nor does it have the '%s' capability", capabilityIfNotParent.Name,
			),
		},
	}
}
