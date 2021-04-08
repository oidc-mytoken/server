package api

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
	APIErrorAuthorizationPending     = APIError{ErrorAuthorizationPending, ""}
	APIErrorConsentDeclined          = APIError{ErrorAccessDenied, "user declined consent"}
	APIErrorNoRefreshToken           = APIError{ErrorOIDC, "Did not receive a refresh token"}
	APIErrorInsufficientCapabilities = APIError{ErrorInsufficientCapabilities, "The provided token does not have the required capability for this operation"}
	APIErrorUsageRestricted          = APIError{ErrorUsageRestricted, "The restrictions of this token does not allow this usage"}
	APIErrorNYI                      = APIError{ErrorNYI, ""}
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
