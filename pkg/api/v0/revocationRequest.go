package api

// RevocationRequest holds the information for a token revocation request
type RevocationRequest struct {
	Token      string `json:"token"` // We don't use model.Token here because we need to revoke a short token differently
	Recursive  bool   `json:"recursive,omitempty"`
	OIDCIssuer string `json:"oidc_issuer,omitempty"`
}
