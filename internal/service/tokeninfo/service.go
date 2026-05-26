// Package tokeninfo provides a service layer for token information operations.
package tokeninfo

import (
	"github.com/gofiber/fiber/v2"

	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
)

// Service is the tokeninfo service singleton.
var Service = &service{}

type service struct{}

func (s *service) makeTokenInfoResponse(rsp interface{}, tokenUpdate *response.MytokenResponse) *model.Response {
	var cake []*fiber.Cookie
	if tokenUpdate != nil {
		cake = []*fiber.Cookie{cookies.MytokenCookie(tokenUpdate.Mytoken)}
	}
	return &model.Response{
		Status:   fiber.StatusOK,
		Response: rsp,
		Cookies:  cake,
	}
}
