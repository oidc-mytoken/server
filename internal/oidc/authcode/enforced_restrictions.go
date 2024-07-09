package authcode

import (
	"fmt"
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/server/httpstatus"
)

func getEnforcedRestrictionTemplate(conf config.EnforcedRestrictionsConf, userInfos map[string]any) (
	string, *model.Response,
) {
	if !conf.Enabled {
		return "", nil
	}

	entitlements, found := userInfos[conf.ClaimName]
	if !found {
		return handleDefaultEnforcedRestrictionsTemplate(conf)
	}

	switch e := entitlements.(type) {
	case string:
		return matchEnforcedRestrictionsTemplate(conf, e)
	case []any:
		strEntitlements := make([]string, len(e))
		for i, entitlement := range e {
			var ok bool
			strEntitlements[i], ok = entitlement.(string)
			if !ok {
				return "", &model.Response{
					Status: httpstatus.StatusOIDPError,
					Response: model.OIDCError(
						"invalid_op_response",
						fmt.Sprintf("cannot understand type of claim '%s'", conf.ClaimName),
					),
				}
			}
		}
		return matchAnyEnforcedRestrictionsTemplate(conf, strEntitlements)
	case []string:
		return matchAnyEnforcedRestrictionsTemplate(conf, e)
	default:
		return "", &model.Response{
			Status: httpstatus.StatusOIDPError,
			Response: model.OIDCError(
				"invalid_op_response",
				fmt.Sprintf("cannot understand type of claim '%s'", conf.ClaimName),
			),
		}
	}
}

func handleDefaultEnforcedRestrictionsTemplate(conf config.EnforcedRestrictionsConf) (string, *model.Response) {
	if conf.ForbidOnDefault {
		return "", &model.Response{
			Status: fiber.StatusForbidden,
			Response: api.Error{
				Error:            api.ErrorStrAccessDenied,
				ErrorDescription: "you do not have the required attributes to use this service",
			},
		}
	}
	return conf.DefaultTemplate, nil
}

func matchEnforcedRestrictionsTemplate(conf config.EnforcedRestrictionsConf, entitlement string) (
	string, *model.Response,
) {
	if template, ok := conf.Mapping[entitlement]; ok {
		return template, nil
	}
	return handleDefaultEnforcedRestrictionsTemplate(conf)
}

func matchAnyEnforcedRestrictionsTemplate(conf config.EnforcedRestrictionsConf, entitlements []string) (
	string, *model.Response,
) {
	for k, v := range conf.Mapping {
		if slices.Contains(entitlements, k) {
			return v, nil
		}
	}
	return handleDefaultEnforcedRestrictionsTemplate(conf)
}
