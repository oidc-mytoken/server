package mytokenlib

import (
	"github.com/oidc-mytoken/server/internal/endpoints/revocation/pkg"
	"github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/httpClient"
)

func (my *MytokenProvider) Revoke(mytoken, oidcIssuer string, recursive bool) error {
	req := pkg.RevocationRequest{
		Token:      mytoken,
		Recursive:  recursive,
		OIDCIssuer: oidcIssuer,
	}
	resp, err := httpClient.Do().R().SetBody(req).SetError(&model.APIError{}).Post(my.RevocationEndpoint)
	if err != nil {
		return newMytokenErrorFromError("error while sending http request", err)
	}
	if e := resp.Error(); e != nil {
		if errRes := e.(*model.APIError); errRes != nil && errRes.Error != "" {
			return &MytokenError{
				err:          errRes.Error,
				errorDetails: errRes.ErrorDescription,
			}
		}
	}
	return nil
}
