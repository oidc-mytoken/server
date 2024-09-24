package authcode

import (
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/userinfo"
	"github.com/oidc-mytoken/server/internal/server/httpstatus"
)

var enforcedRestrictionsClaimSourcesUserInfoKeys = []string{
	"op",
	"issuer",
	"default",
	"userinfo",
}

func getEnforcedRestrictionTemplate(
	conf config.EnforcedRestrictionsConf,
	userInfos map[string]any, at string,
) (
	string, bool, *model.Response,
) {
	if !conf.Enabled {
		return "", false, nil
	}

	var claimValue any
	var found bool
	for endpoint, claimName := range conf.ClaimSources {
		if slices.Contains(enforcedRestrictionsClaimSourcesUserInfoKeys, endpoint) {
			claimValue, found = userInfos[claimName]
		} else {
			userAttributes, errRes, err := userinfo.Get(endpoint, at)
			if err != nil || errRes != nil || userAttributes == nil {
				continue
			}
			claimValue, found = userAttributes[claimName]
		}
		if found {
			break
		}
	}
	if !found {
		return handleDefaultEnforcedRestrictionsTemplate(conf)
	}

	switch e := claimValue.(type) {
	case string:
		return matchEnforcedRestrictionsTemplate(conf, e)
	case []any:
		strEntitlements := make([]string, len(e))
		for i, entitlement := range e {
			var ok bool
			strEntitlements[i], ok = entitlement.(string)
			if !ok {
				return "", false, &model.Response{
					Status: httpstatus.StatusOIDPError,
					Response: model.OIDCError(
						"invalid_claim_source_response",
						"cannot understand claim type",
					),
				}
			}
		}
		return matchAnyEnforcedRestrictionsTemplate(conf, strEntitlements)
	case []string:
		return matchAnyEnforcedRestrictionsTemplate(conf, e)
	default:
		return "", false, &model.Response{
			Status: httpstatus.StatusOIDPError,
			Response: model.OIDCError(
				"invalid_claim_source_response",
				"cannot understand claim type",
			),
		}
	}
}

func handleDefaultEnforcedRestrictionsTemplate(conf config.EnforcedRestrictionsConf) (string, bool, *model.Response) {
	if conf.ForbidOnDefault {
		return "", true, &model.Response{
			Status: fiber.StatusForbidden,
			Response: api.Error{
				Error:            api.ErrorStrAccessDenied,
				ErrorDescription: "you do not have the required attributes to use this service",
			},
		}
	}
	return conf.DefaultTemplate, false, nil
}

func matchEnforcedRestrictionsTemplate(conf config.EnforcedRestrictionsConf, entitlement string) (
	string, bool, *model.Response,
) {
	if template, ok := conf.Mapping[entitlement]; ok {
		return template, false, nil
	}
	return handleDefaultEnforcedRestrictionsTemplate(conf)
}

func matchAnyEnforcedRestrictionsTemplate(conf config.EnforcedRestrictionsConf, entitlements []string) (
	string, bool, *model.Response,
) {
	for k, v := range conf.Mapping {
		if slices.Contains(entitlements, k) {
			return v, false, nil
		}
	}
	return handleDefaultEnforcedRestrictionsTemplate(conf)
}
