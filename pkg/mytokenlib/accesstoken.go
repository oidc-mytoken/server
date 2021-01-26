package mytokenlib

import (
	"strings"

	"github.com/zachmann/mytoken/internal/server/endpoints/token/access/pkg"
	"github.com/zachmann/mytoken/internal/supertoken/token"
	"github.com/zachmann/mytoken/pkg/model"
	"github.com/zachmann/mytoken/shared/httpClient"
)

func (my *Mytoken) GetAccessToken(superToken, oidcIssuer string, scopes []string, audiences []string, comment string) (string, error) {
	req := pkg.AccessTokenRequest{
		Issuer:     oidcIssuer,
		GrantType:  model.GrantTypeSuperToken,
		SuperToken: token.Token(superToken),
		Scope:      strings.Join(scopes, " "),
		Audience:   strings.Join(audiences, " "),
		Comment:    comment,
	}
	resp, err := httpClient.Do().R().SetBody(req).SetResult(&pkg.AccessTokenResponse{}).SetError(&model.APIError{}).Post(my.AccessTokenEndpoint)
	if err != nil {
		return "", newMytokenErrorFromError("error while sending http request", err)
	}
	if e := resp.Error(); e != nil {
		if errRes := e.(*model.APIError); errRes != nil && len(errRes.Error) > 0 {
			return "", &MytokenError{
				err:          errRes.Error,
				errorDetails: errRes.ErrorDescription,
			}
		}
	}
	atRes, ok := resp.Result().(*pkg.AccessTokenResponse)
	if !ok {
		return "", &MytokenError{
			err: "unexpected response from mytoken server",
		}
	}
	return atRes.AccessToken, nil
}
