package api

var AllGrantTypes = [...]string{GrantTypeMytoken, GrantTypeOIDCFlow, GrantTypePollingCode, GrantTypeAccessToken, GrantTypePrivateKeyJWT, GrantTypeTransferCode}

// GrantTypes
const (
	GrantTypeMytoken       = "mytoken"
	GrantTypeOIDCFlow      = "oidc_flow"
	GrantTypePollingCode   = "polling_code"
	GrantTypeAccessToken   = "access_token"
	GrantTypePrivateKeyJWT = "private_key_jwt"
	GrantTypeTransferCode  = "transfer_code"
)
