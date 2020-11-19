package model

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type APIError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

var ResponseNYI = Response{Status: fiber.StatusNotImplemented, Response: APIErrorNYI}

// Predefined errors
var (
	APIErrorUnknownIssuer            = APIError{ErrorInvalidRequest, "The provided issuer is not supported"}
	APIErrorStateMismatch            = APIError{ErrorInvalidRequest, "State mismatched"}
	APIErrorUnsupportedOIDCFlow      = APIError{ErrorInvalidGrant, "Unsupported oidc_flow"}
	APIErrorUnsupportedGrantType     = APIError{ErrorInvalidGrant, "Unsupported grant_type"}
	APIErrorBadPollingCode           = APIError{ErrorAccessDenied, "Bad polling_code"}
	APIErrorPollingCodeExpired       = APIError{ErrorExpiredToken, "polling_code is expired"}
	APIErrorAuthorizationPending     = ErrorWithoutDescription(ErrorAuthorizationPending)
	APIErrorNoRefreshToken           = APIError{ErrorOIDC, "Did not receive a refresh token"}
	APIErrorInsufficientCapabilities = APIError{ErrorInsufficientCapabilities, "The provided token does not have the required capability for this operation"}
	APIErrorUsageRestricted          = APIError{ErrorUsageRestricted, "The restrictions of this token does not allow this usage"}
	APIErrorNYI                      = ErrorWithoutDescription(ErrorNYI)
)

// Predefined OAuth2/OIDC errors
const (
	ErrorInvalidRequest       = "invalid_request"
	ErrorInvalidClient        = "invalid_client"
	ErrorInvalidGrant         = "invalid_grant"
	ErrorUnauthorizedClient   = "unauthorized_client"
	ErrorUnsupportedGrantType = "unsupported_grant_type"
	ErrorInvalidScope         = "invalid_scope"
	ErrorInvalidToken         = "invalid_token"
	ErrorInsufficientScope    = "insufficient_scope"
	ErrorExpiredToken         = "expired_token"
	ErrorAccessDenied         = "access_denied"
	ErrorAuthorizationPending = "authorization_pending"
)

// Additional Mytoken errors
const (
	ErrorInternal                 = "internal_server_error"
	ErrorOIDC                     = "oidc_error"
	ErrorNYI                      = "not_yet_implemented"
	ErrorInsufficientCapabilities = "insufficient_capabilities"
	ErrorUsageRestricted          = "usage_restricted"
)

func InternalServerError(errorDescription string) APIError {
	return APIError{
		Error:            ErrorInternal,
		ErrorDescription: errorDescription,
	}
}

func OIDCError(oidcError, oidcErrorDescription string) APIError {
	error := oidcError
	if oidcErrorDescription != "" {
		error = fmt.Sprintf("%s: %s", oidcError, oidcErrorDescription)
	}
	return APIError{
		Error:            ErrorOIDC,
		ErrorDescription: error,
	}
}

func OIDCErrorFromBody(body []byte) (error APIError, ok bool) {
	bodyError := APIError{}
	if err := json.Unmarshal(body, &bodyError); err != nil {
		return
	}
	error = OIDCError(bodyError.Error, bodyError.ErrorDescription)
	ok = true
	return
}

func BadRequestError(errorDescription string) APIError {
	return APIError{
		Error:            ErrorInvalidRequest,
		ErrorDescription: errorDescription,
	}
}

func ErrorWithoutDescription(error string) APIError {
	return APIError{
		Error: error,
	}
}

func ErrorWithErrorDescription(error string, err error) APIError {
	return APIError{
		Error:            error,
		ErrorDescription: err.Error(),
	}
}
