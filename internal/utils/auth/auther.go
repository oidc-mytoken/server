package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	dbhelper "github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/utils/ctxUtils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	model2 "github.com/oidc-mytoken/server/shared/model"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
	"github.com/oidc-mytoken/server/shared/mytoken/universalmytoken"
)

// RequireGrantType checks that the passed model.GrantType are the same, and returns an error model.Response if not
func RequireGrantType(want, got model2.GrantType) *model.Response {
	if got != want {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnsupportedGrantType,
		}
	}
	log.Trace("Checked grant type")
	return nil
}

// RequireMytoken checks the passed universalmytoken.UniversalMytoken and if needed other request parameters like
// authorization header and cookie value for a mytoken string. The mytoken string is parsed and if not valid an error
// model.Response is returned.
func RequireMytoken(reqToken *universalmytoken.UniversalMytoken, ctx *fiber.Ctx) (*mytoken.Mytoken, *model.Response) {
	if reqToken.JWT == "" {
		t, found := ctxUtils.GetMytoken(ctx)
		if t == nil {
			errDesc := "no mytoken found in request"
			if found {
				errDesc = "token not valid"
			}
			return nil, &model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: model2.InvalidTokenError(errDesc),
			}
		}
		*reqToken = *t
	}

	mt, err := mytoken.ParseJWT(reqToken.JWT)
	if err != nil {
		return nil, &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model2.InvalidTokenError(errorfmt.Error(err)),
		}
	}
	log.Trace("Parsed mytoken")
	return mt, nil
}

// RequireMytokenNotRevoked checks that the passed mytoken.Mytoken was not revoked, if it was an error model.Response is
// returned.
func RequireMytokenNotRevoked(tx *sqlx.Tx, mt *mytoken.Mytoken) *model.Response {
	revoked, dbErr := dbhelper.CheckTokenRevoked(tx, mt.ID, mt.SeqNo, mt.Rotation)
	if dbErr != nil {
		log.Errorf("%s", errorfmt.Full(dbErr))
		return model.ErrorToInternalServerErrorResponse(dbErr)
	}
	if revoked {
		return &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model2.InvalidTokenError(""),
		}
	}
	log.Trace("Checked mytoken not revoked")
	return nil
}

// RequireValidMytoken checks the passed universalmytoken.UniversalMytoken and if needed other request parameters like
// authorization header and cookie value for a mytoken string. The mytoken string is parsed and if not valid an error
// model.Response is returned. RequireValidMytoken also asserts that the mytoken.Mytoken was not revoked.
func RequireValidMytoken(tx *sqlx.Tx, reqToken *universalmytoken.UniversalMytoken, ctx *fiber.Ctx) (*mytoken.Mytoken, *model.Response) {
	mt, errRes := RequireMytoken(reqToken, ctx)
	if errRes != nil {
		return nil, errRes
	}
	return mt, RequireMytokenNotRevoked(tx, mt)
}

// RequireMatchingIssuer checks that the OIDC issuer from a mytoken is the same as the issuer string in a request (if
// given). RequireMatchingIssuer also checks that the issuer is valid for this mytoken instance.
func RequireMatchingIssuer(mtOIDCIssuer string, requestIssuer *string) (*config.ProviderConf, *model.Response) {
	if *requestIssuer == "" {
		*requestIssuer = mtOIDCIssuer
		log.Trace("Checked issuer (was not given)")
	}
	if *requestIssuer != mtOIDCIssuer {
		return nil, &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model2.BadRequestError("token not for specified issuer"),
		}
	}
	provider, ok := config.Get().ProviderByIssuer[*requestIssuer]
	if !ok {
		return nil, &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnknownIssuer,
		}
	}
	log.Trace("Checked issuer")
	return provider, nil
}

// RequireCapability checks that the passed mytoken.Mytoken has the required api.Capability and returns an error
// model.Response if not
func RequireCapability(capability api.Capability, mt *mytoken.Mytoken) *model.Response {
	if !mt.Capabilities.Has(capability) {
		return &model.Response{
			Status:   fiber.StatusForbidden,
			Response: api.ErrorInsufficientCapabilities,
		}
	}
	log.Trace("Checked capability")
	return nil
}

func requireUseableRestriction(tx *sqlx.Tx, mt *mytoken.Mytoken, ip string, scopes, auds []string, at bool) (*restrictions.Restriction, *model.Response) {
	if len(mt.Restrictions) == 0 {
		return nil, nil
	}
	getUseableRestrictions := mt.Restrictions.GetValidForOther
	if at {
		getUseableRestrictions = mt.Restrictions.GetValidForAT
	}
	// WithScopes and WithAudience don't tighten the restrictions if nil is passed
	useableRestrictions := getUseableRestrictions(tx, ip, mt.ID).WithScopes(scopes).WithAudiences(auds)
	if len(useableRestrictions) == 0 {
		return nil, &model.Response{
			Status:   fiber.StatusForbidden,
			Response: api.ErrorUsageRestricted,
		}
	}
	log.Trace("Checked mytoken restrictions")
	return &useableRestrictions[0], nil
}

// RequireUsableRestriction checks that the mytoken.Mytoken's restrictions allow the usage
func RequireUsableRestriction(tx *sqlx.Tx, mt *mytoken.Mytoken, ip string, scopes, auds []string, capability api.Capability) (*restrictions.Restriction, *model.Response) {
	return requireUseableRestriction(tx, mt, ip, scopes, auds, capability == api.CapabilityAT)
}

// RequireUsableRestrictionAT checks that the mytoken.Mytoken's restrictions allow the AT usage
func RequireUsableRestrictionAT(tx *sqlx.Tx, mt *mytoken.Mytoken, ip string, scopes, auds []string) (*restrictions.Restriction, *model.Response) {
	return requireUseableRestriction(tx, mt, ip, scopes, auds, true)
}

// RequireUsableRestrictionOther checks that the mytoken.Mytoken's restrictions allow the non-AT usage
func RequireUsableRestrictionOther(tx *sqlx.Tx, mt *mytoken.Mytoken, ip string, scopes, auds []string) (*restrictions.Restriction, *model.Response) {
	return requireUseableRestriction(tx, mt, ip, scopes, auds, false)
}

// CheckCapabilityAndRestriction checks the mytoken.Mytoken's capability and restrictions
func CheckCapabilityAndRestriction(tx *sqlx.Tx, mt *mytoken.Mytoken, ip string, scopes, auds []string, capability api.Capability) (*restrictions.Restriction, *model.Response) {
	if errRes := RequireCapability(capability, mt); errRes != nil {
		return nil, errRes
	}
	return RequireUsableRestriction(tx, mt, ip, scopes, auds, capability)
}
