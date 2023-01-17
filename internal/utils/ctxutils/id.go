package ctxutils

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type idUnmarshal struct {
	ID uuid.UUID `json:"id"`
}

// GetID returns the id for a fiber.Ctx by checking the url param, query as well as the request body (json)
func GetID(ctx *fiber.Ctx) (uuid.UUID, error) {
	id := ctx.Params("id")
	if id != "" {
		return uuid.FromString(id)
	}
	id = ctx.Query("id")
	if id != "" {
		return uuid.FromString(id)
	}
	i := idUnmarshal{}
	err := errors.WithStack(json.Unmarshal(ctx.Body(), &i))
	if err != nil {
		return uuid.Nil, err
	}
	return i.ID, nil
}
