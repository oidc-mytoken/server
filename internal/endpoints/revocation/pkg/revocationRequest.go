package pkg

// RevocationRequest holds the information for a token revocation request
type RevocationRequest struct {
	Token      string `json:"token"`
	Recursive  bool   `json:"recursive"`
	OIDCIssuer string `json:"oidc_issuer"`
}
