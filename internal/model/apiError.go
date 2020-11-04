package model

type APIError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// Predefined errors
var (
	APIErrorUnknownIssuer = APIError{ErrorInvalidRequest, "The provided issuer is not supported"}
	APIErrorStateMismatch = APIError{ErrorInvalidRequest, "State mismatched"}
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
)

// Additional Mytoken errors
const (
	ErrorInternal = "internal_server_error"
)

func InternalServerError(errorDescription string) APIError {
	return APIError{
		Error:            ErrorInternal,
		ErrorDescription: errorDescription,
	}
}
