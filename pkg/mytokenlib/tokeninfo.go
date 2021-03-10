package mytokenlib

import (
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/httpClient"
	"github.com/oidc-mytoken/server/shared/supertoken/token"
)

func (my *Mytoken) TokeninfoIntrospect(superToken string) (*pkg.TokeninfoIntrospectResponse, error) {
	req := pkg.TokenInfoRequest{
		Action:     model.TokeninfoActionIntrospect,
		SuperToken: token.Token(superToken),
	}
	resp, err := httpClient.Do().R().SetBody(req).SetResult(&pkg.TokeninfoIntrospectResponse{}).SetError(&model.APIError{}).Post(my.TokeninfoEndpoint)
	if err != nil {
		return nil, newMytokenErrorFromError("error while sending http request", err)
	}
	if eRes := resp.Error(); eRes != nil {
		if errRes := eRes.(*model.APIError); errRes != nil && errRes.Error != "" {
			return nil, &MytokenError{
				err:          errRes.Error,
				errorDetails: errRes.ErrorDescription,
			}
		}
	}
	res, ok := resp.Result().(*pkg.TokeninfoIntrospectResponse)
	if !ok {
		return nil, &MytokenError{
			err: "unexpected response from mytoken server",
		}
	}
	return res, nil
}
func (my *Mytoken) TokeninfoHistory(superToken string) (*pkg.TokeninfoHistoryResponse, error) {
	req := pkg.TokenInfoRequest{
		Action:     model.TokeninfoActionEventHistory,
		SuperToken: token.Token(superToken),
	}
	resp, err := httpClient.Do().R().SetBody(req).SetResult(&pkg.TokeninfoHistoryResponse{}).SetError(&model.APIError{}).Post(my.TokeninfoEndpoint)
	if err != nil {
		return nil, newMytokenErrorFromError("error while sending http request", err)
	}
	if eRes := resp.Error(); eRes != nil {
		if errRes := eRes.(*model.APIError); errRes != nil && errRes.Error != "" {
			return nil, &MytokenError{
				err:          errRes.Error,
				errorDetails: errRes.ErrorDescription,
			}
		}
	}
	res, ok := resp.Result().(*pkg.TokeninfoHistoryResponse)
	if !ok {
		return nil, &MytokenError{
			err: "unexpected response from mytoken server",
		}
	}
	return res, nil
}
func (my *Mytoken) TokeninfoSubtokens(superToken string) (*pkg.TokeninfoTreeResponse, error) {
	req := pkg.TokenInfoRequest{
		Action:     model.TokeninfoActionSubtokenTree,
		SuperToken: token.Token(superToken),
	}
	resp, err := httpClient.Do().R().SetBody(req).SetResult(&pkg.TokeninfoTreeResponse{}).SetError(&model.APIError{}).Post(my.TokeninfoEndpoint)
	if err != nil {
		return nil, newMytokenErrorFromError("error while sending http request", err)
	}
	if eRes := resp.Error(); eRes != nil {
		if errRes := eRes.(*model.APIError); errRes != nil && errRes.Error != "" {
			return nil, &MytokenError{
				err:          errRes.Error,
				errorDetails: errRes.ErrorDescription,
			}
		}
	}
	res, ok := resp.Result().(*pkg.TokeninfoTreeResponse)
	if !ok {
		return nil, &MytokenError{
			err: "unexpected response from mytoken server",
		}
	}
	return res, nil
}
func (my *Mytoken) TokeninfoListSuperTokens(superToken string) (*pkg.TokeninfoListResponse, error) {
	req := pkg.TokenInfoRequest{
		Action:     model.TokeninfoActionListSuperTokens,
		SuperToken: token.Token(superToken),
	}
	resp, err := httpClient.Do().R().SetBody(req).SetResult(&pkg.TokeninfoListResponse{}).SetError(&model.APIError{}).Post(my.TokeninfoEndpoint)
	if err != nil {
		return nil, newMytokenErrorFromError("error while sending http request", err)
	}
	if eRes := resp.Error(); eRes != nil {
		if errRes := eRes.(*model.APIError); errRes != nil && errRes.Error != "" {
			return nil, &MytokenError{
				err:          errRes.Error,
				errorDetails: errRes.ErrorDescription,
			}
		}
	}
	res, ok := resp.Result().(*pkg.TokeninfoListResponse)
	if !ok {
		return nil, &MytokenError{
			err: "unexpected response from mytoken server",
		}
	}
	return res, nil
}
