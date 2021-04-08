package mytokenlib

import (
	"github.com/oidc-mytoken/server/pkg/api/v0"
	"github.com/oidc-mytoken/server/shared/httpClient"
	"github.com/oidc-mytoken/server/shared/utils"
)

type MytokenProvider struct {
	api.MytokenConfiguration
}

func NewMytokenProvider(url string) (*MytokenProvider, error) {
	configEndpoint := utils.CombineURLPath(url, "/.well-known/mytoken-configuration")
	resp, err := httpClient.Do().R().SetResult(&api.MytokenConfiguration{}).SetError(&api.APIError{}).Get(configEndpoint)
	if err != nil {
		return nil, newMytokenErrorFromError("could not connect to mytoken instance", err)
	}
	if e := resp.Error(); e != nil {
		if errRes := e.(*api.APIError); errRes != nil && errRes.Error != "" {
			return nil, &MytokenError{
				err:          errRes.Error,
				errorDetails: errRes.ErrorDescription,
			}
		}
	}
	config, ok := resp.Result().(*api.MytokenConfiguration)
	if !ok {
		return nil, &MytokenError{
			err: unexpectedResponse,
		}
	}
	return &MytokenProvider{
		MytokenConfiguration: *config,
	}, nil
}
