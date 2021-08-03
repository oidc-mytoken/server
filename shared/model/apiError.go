package model

import (
	"encoding/json"
	"fmt"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
)

// InternalServerError creates an Error for internal server errors
func InternalServerError(errorDescription string) api.Error {
	return api.Error{
		Error:            api.ErrorStrInternal,
		ErrorDescription: errorDescription,
	}
}

// OIDCError creates an Error for oidc related errors
func OIDCError(oidcError, oidcErrorDescription string) api.Error {
	err := oidcError
	if oidcErrorDescription != "" {
		err = fmt.Sprintf("%s: %s", oidcError, oidcErrorDescription)
	}
	return api.Error{
		Error:            api.ErrorStrOIDC,
		ErrorDescription: err,
	}
}

// OIDCErrorFromBody creates an Error for oidc related errors from the response of an oidc provider
func OIDCErrorFromBody(body []byte) (apiError api.Error, ok bool) {
	bodyError := api.Error{}
	if err := json.Unmarshal(body, &bodyError); err != nil {
		return
	}
	apiError = OIDCError(bodyError.Error, bodyError.ErrorDescription)
	ok = true
	return
}

// BadRequestError creates an Error for bad request errors
func BadRequestError(errorDescription string) api.Error {
	return api.Error{
		Error:            api.ErrorStrInvalidRequest,
		ErrorDescription: errorDescription,
	}
}

// InvalidTokenError creates an Error for invalid token errors
func InvalidTokenError(errorDescription string) api.Error {
	return api.Error{
		Error:            api.ErrorStrInvalidToken,
		ErrorDescription: errorDescription,
	}
}

// ErrorWithoutDescription creates an Error from an error string
func ErrorWithoutDescription(err string) api.Error {
	return api.Error{
		Error: err,
	}
}

// ErrorWithErrorDescription creates an Error from an error string and golang error
func ErrorWithErrorDescription(e string, err error) api.Error {
	return api.Error{
		Error:            e,
		ErrorDescription: errorfmt.Error(err),
	}
}
