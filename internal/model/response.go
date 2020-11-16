package model

import (
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

type Response struct {
	Status   int
	Response interface{}
	Cookies  []*fiber.Cookie
}

func (r *Response) Send(ctx *fiber.Ctx) error {
	if fasthttp.StatusCodeIsRedirect(r.Status) {
		ctx.Redirect(r.Response.(string), r.Status)
	}
	if r.Cookies != nil && len(r.Cookies) > 0 {
		for _, c := range r.Cookies {
			ctx.Cookie(c)
		}
	}
	return ctx.Status(r.Status).JSON(r.Response)
}

func ErrorToInternalServerErrorResponse(err error) Response {
	return Response{
		Status:   fiber.StatusInternalServerError,
		Response: InternalServerError(err.Error()),
	}
}
