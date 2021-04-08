package model

import (
	"encoding/json"
	"fmt"

	"github.com/oidc-mytoken/server/pkg/api/v0"
)

// InternalServerError creates an APIError for internal server errors
func InternalServerError(errorDescription string) api.APIError {
	return api.APIError{
		Error:            api.ErrorInternal,
		ErrorDescription: errorDescription,
	}
}

// OIDCError creates an APIError for oidc related errors
func OIDCError(oidcError, oidcErrorDescription string) api.APIError {
	err := oidcError
	if oidcErrorDescription != "" {
		err = fmt.Sprintf("%s: %s", oidcError, oidcErrorDescription)
	}
	return api.APIError{
		Error:            api.ErrorOIDC,
		ErrorDescription: err,
	}
}

// OIDCErrorFromBody creates an APIError for oidc related errors from the response of an oidc provider
func OIDCErrorFromBody(body []byte) (apiError api.APIError, ok bool) {
	bodyError := api.APIError{}
	if err := json.Unmarshal(body, &bodyError); err != nil {
		return
	}
	apiError = OIDCError(bodyError.Error, bodyError.ErrorDescription)
	ok = true
	return
}

// BadRequestError creates an APIError for bad request errors
func BadRequestError(errorDescription string) api.APIError {
	return api.APIError{
		Error:            api.ErrorInvalidRequest,
		ErrorDescription: errorDescription,
	}
}

// InvalidTokenError creates an APIError for invalid token errors
func InvalidTokenError(errorDescription string) api.APIError {
	return api.APIError{
		Error:            api.ErrorInvalidToken,
		ErrorDescription: errorDescription,
	}
}

// ErrorWithoutDescription creates an APIError from an error string
func ErrorWithoutDescription(err string) api.APIError {
	return api.APIError{
		Error: err,
	}
}

// ErrorWithErrorDescription creates an APIError from an error string and golang error
func ErrorWithErrorDescription(e string, err error) api.APIError {
	return api.APIError{
		Error:            e,
		ErrorDescription: err.Error(),
	}
}
