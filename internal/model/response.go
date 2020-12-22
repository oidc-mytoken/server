package model

import (
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

// Response models a http server response
type Response struct {
	// The Http Status code of the response
	Status int
	// The response body, will be marshalled as json
	Response interface{}
	// Cookies that should be set
	Cookies []*fiber.Cookie
}

// Send sends this response using the passed fiber.Ctx
func (r Response) Send(ctx *fiber.Ctx) error {
	if r.Cookies != nil && len(r.Cookies) > 0 {
		for _, c := range r.Cookies {
			ctx.Cookie(c)
		}
	}
	if fasthttp.StatusCodeIsRedirect(r.Status) {
		return ctx.Redirect(r.Response.(string), r.Status)
	}
	return ctx.Status(r.Status).JSON(r.Response)
}

// ErrorToInternalServerErrorResponse creates an internal server error response from a golang error
func ErrorToInternalServerErrorResponse(err error) *Response {
	return &Response{
		Status:   fiber.StatusInternalServerError,
		Response: InternalServerError(err.Error()),
	}
}

// ErrorToBadRequestErrorResponse creates a bad request error response from a golang error
func ErrorToBadRequestErrorResponse(err error) *Response {
	return &Response{
		Status:   fiber.StatusBadRequest,
		Response: BadRequestError(err.Error()),
	}
}
