package model

import (
	"encoding/json"
	"fmt"
)

// APIError is an error object that is returned on the api when an error occurs
type APIError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// Predefined errors
var (
	APIErrorUnknownIssuer            = APIError{ErrorInvalidRequest, "The provided issuer is not supported"}
	APIErrorStateMismatch            = APIError{ErrorInvalidRequest, "State mismatched"}
	APIErrorUnsupportedOIDCFlow      = APIError{ErrorInvalidGrant, "Unsupported oidc_flow"}
	APIErrorUnsupportedGrantType     = APIError{ErrorInvalidGrant, "Unsupported grant_type"}
	APIErrorBadTransferCode          = APIError{ErrorInvalidToken, "Bad polling or transfer code"}
	APIErrorTransferCodeExpired      = APIError{ErrorExpiredToken, "polling or transfer code is expired"}
	APIErrorAuthorizationPending     = ErrorWithoutDescription(ErrorAuthorizationPending)
	APIErrorConsentDeclined          = APIError{ErrorAccessDenied, "user declined consent"}
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

// InternalServerError creates an APIError for internal server errors
func InternalServerError(errorDescription string) APIError {
	return APIError{
		Error:            ErrorInternal,
		ErrorDescription: errorDescription,
	}
}

// OIDCError creates an APIError for oidc related errors
func OIDCError(oidcError, oidcErrorDescription string) APIError {
	err := oidcError
	if oidcErrorDescription != "" {
		err = fmt.Sprintf("%s: %s", oidcError, oidcErrorDescription)
	}
	return APIError{
		Error:            ErrorOIDC,
		ErrorDescription: err,
	}
}

// OIDCErrorFromBody creates an APIError for oidc related errors from the response of an oidc provider
func OIDCErrorFromBody(body []byte) (error APIError, ok bool) {
	bodyError := APIError{}
	if err := json.Unmarshal(body, &bodyError); err != nil {
		return
	}
	error = OIDCError(bodyError.Error, bodyError.ErrorDescription)
	ok = true
	return
}

// BadRequestError creates an APIError for bad request errors
func BadRequestError(errorDescription string) APIError {
	return APIError{
		Error:            ErrorInvalidRequest,
		ErrorDescription: errorDescription,
	}
}

// InvalidTokenError creates an APIError for invalid token errors
func InvalidTokenError(errorDescription string) APIError {
	return APIError{
		Error:            ErrorInvalidToken,
		ErrorDescription: errorDescription,
	}
}

// ErrorWithoutDescription creates an APIError from an error string
func ErrorWithoutDescription(err string) APIError {
	return APIError{
		Error: err,
	}
}

// ErrorWithErrorDescription creates an APIError from an error string and golang error
func ErrorWithErrorDescription(e string, err error) APIError {
	return APIError{
		Error:            e,
		ErrorDescription: err.Error(),
	}
}
