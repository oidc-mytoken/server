package mytokenlib

import (
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/httpClient"
	"github.com/oidc-mytoken/server/shared/mytoken/token"
)

func (my *MytokenProvider) TokeninfoIntrospect(mytoken string) (*pkg.TokeninfoIntrospectResponse, error) {
	req := pkg.TokenInfoRequest{
		Action:  model.TokeninfoActionIntrospect,
		Mytoken: token.Token(mytoken),
	}
	resp, err := httpClient.Do().R().SetBody(req).SetResult(&pkg.TokeninfoIntrospectResponse{}).SetError(&model.APIError{}).Post(my.TokeninfoEndpoint)
	if err != nil {
		return nil, newMytokenErrorFromError(errorWhileHttp, err)
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
			err: unexpectedResponse,
		}
	}
	return res, nil
}
func (my *MytokenProvider) TokeninfoHistory(mytoken string) (*pkg.TokeninfoHistoryResponse, error) {
	req := pkg.TokenInfoRequest{
		Action:  model.TokeninfoActionEventHistory,
		Mytoken: token.Token(mytoken),
	}
	resp, err := httpClient.Do().R().SetBody(req).SetResult(&pkg.TokeninfoHistoryResponse{}).SetError(&model.APIError{}).Post(my.TokeninfoEndpoint)
	if err != nil {
		return nil, newMytokenErrorFromError(errorWhileHttp, err)
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
			err: unexpectedResponse,
		}
	}
	return res, nil
}
func (my *MytokenProvider) TokeninfoSubtokens(mytoken string) (*pkg.TokeninfoTreeResponse, error) {
	req := pkg.TokenInfoRequest{
		Action:  model.TokeninfoActionSubtokenTree,
		Mytoken: token.Token(mytoken),
	}
	resp, err := httpClient.Do().R().SetBody(req).SetResult(&pkg.TokeninfoTreeResponse{}).SetError(&model.APIError{}).Post(my.TokeninfoEndpoint)
	if err != nil {
		return nil, newMytokenErrorFromError(errorWhileHttp, err)
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
			err: unexpectedResponse,
		}
	}
	return res, nil
}
func (my *MytokenProvider) TokeninfoListMytokens(mytoken string) (*pkg.TokeninfoListResponse, error) {
	req := pkg.TokenInfoRequest{
		Action:  model.TokeninfoActionListMytokens,
		Mytoken: token.Token(mytoken),
	}
	resp, err := httpClient.Do().R().SetBody(req).SetResult(&pkg.TokeninfoListResponse{}).SetError(&model.APIError{}).Post(my.TokeninfoEndpoint)
	if err != nil {
		return nil, newMytokenErrorFromError(errorWhileHttp, err)
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
			err: unexpectedResponse,
		}
	}
	return res, nil
}
