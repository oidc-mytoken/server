package mytokenlib

import (
	api "github.com/oidc-mytoken/server/pkg/api/v0"
	"github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/httpClient"
)

func (my *MytokenProvider) GetAccessToken(mytoken, oidcIssuer string, scopes, audiences []string, comment string) (string, error) {
	req := NewAccessTokenRequest(oidcIssuer, mytoken, scopes, audiences, comment)
	resp, err := httpClient.Do().R().SetBody(req).SetResult(&api.AccessTokenResponse{}).SetError(&model.APIError{}).Post(my.AccessTokenEndpoint)
	if err != nil {
		return "", newMytokenErrorFromError("error while sending http request", err)
	}
	if e := resp.Error(); e != nil {
		if errRes := e.(*model.APIError); errRes != nil && errRes.Error != "" {
			return "", &MytokenError{
				err:          errRes.Error,
				errorDetails: errRes.ErrorDescription,
			}
		}
	}
	atRes, ok := resp.Result().(*api.AccessTokenResponse)
	if !ok {
		return "", &MytokenError{
			err: unexpectedResponse,
		}
	}
	return atRes.AccessToken, nil
}
