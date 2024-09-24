package userinfo

import (
	"github.com/oidc-mytoken/utils/httpclient"
	"github.com/oidc-mytoken/utils/utils/jwtutils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/oidcreqres"
)

// Get obtains the userinfo response from the passed endpoint
func Get(
	endpoint string, at string,
) (map[string]any, *oidcreqres.OIDCErrorResponse, error) {

	httpRes, err := httpclient.Do().R().
		SetAuthToken(at).
		SetResult(make(map[string]any)).
		SetError(&oidcreqres.OIDCErrorResponse{}).
		Get(endpoint)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	if errRes, ok := httpRes.Error().(*oidcreqres.OIDCErrorResponse); ok && errRes != nil && errRes.Error != "" {
		errRes.Status = httpRes.RawResponse.StatusCode
		return nil, errRes, nil
	}
	res, ok := httpRes.Result().(map[string]any)
	if !ok {
		return nil, nil, errors.New("could not unmarshal userinfo response")
	}
	return res, nil, nil
}

// GetFromProvider obtains the userinfo response from the model.Provider's
// userinfo endpoint
func GetFromProvider(
	provider model.Provider, at string,
) (map[string]any, *oidcreqres.OIDCErrorResponse, error) {
	return Get(provider.Endpoints().Userinfo, at)
}

func getNonNilUserInfoMap(provider model.Provider, at string) map[string]any {
	userinfoRes, errRes, err := GetFromProvider(provider, at)
	if err == nil && errRes == nil {
		return userinfoRes
	}
	return map[string]any{}
}

func getNonNilJWTMap(rlog log.Ext1FieldLogger, token string) map[string]any {
	attrs := jwtutils.GetFromJWT(rlog, token)
	if attrs == nil {
		attrs = map[string]any{}
	}
	return attrs
}

// GetUserAttributes returns user attributes for the passed claim names by searching the id token, JWT AT,
// and userinfo endpoint
func GetUserAttributes(
	rlog log.Ext1FieldLogger, oidcTokenRes *oidcreqres.OIDCTokenResponse, provider model.Provider,
	attributes ...string,
) map[string]any {
	var atTokenAttrs map[string]any
	var userInfoAttrs map[string]any
	idTokenAttrs := getNonNilJWTMap(rlog, oidcTokenRes.IDToken)

	finalAttrs := make(map[string]any, len(attributes))
	for _, attr := range attributes {
		if v, ok := idTokenAttrs[attr]; ok {
			finalAttrs[attr] = v
			continue
		}
		if atTokenAttrs == nil {
			atTokenAttrs = getNonNilJWTMap(rlog, oidcTokenRes.AccessToken)
		}
		if v, ok := atTokenAttrs[attr]; ok {
			finalAttrs[attr] = v
			continue
		}
		if userInfoAttrs == nil {
			userInfoAttrs = getNonNilUserInfoMap(provider, oidcTokenRes.AccessToken)
		}
		if v, ok := userInfoAttrs[attr]; ok {
			finalAttrs[attr] = v
			// continue
		}
	}
	return finalAttrs
}
