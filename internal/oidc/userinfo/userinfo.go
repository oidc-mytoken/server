package userinfo

import (
	"github.com/oidc-mytoken/utils/httpclient"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/oidcreqres"
)

// Get obtains the userinfo response from the model.Provider's userinfo endpoint
func Get(
	provider model.Provider, at string,
) (map[string]any, *oidcreqres.OIDCErrorResponse, error) {

	httpRes, err := httpclient.Do().R().
		SetAuthToken(at).
		SetResult(make(map[string]any)).
		SetError(&oidcreqres.OIDCErrorResponse{}).
		Get(provider.Endpoints().Userinfo)
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
