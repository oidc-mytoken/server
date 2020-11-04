package model

import "github.com/gofiber/fiber/v2"

type Response struct {
	Status   int
	Response interface{}
}

func (r *Response) Send(ctx *fiber.Ctx) error {
	return ctx.Status(r.Status).JSON(r.Response)
}

func ErrorToInternalServerErrorResponse(err error) Response {
	return Response{
		Status:   fiber.StatusInternalServerError,
		Response: InternalServerError(err.Error()),
	}
}
