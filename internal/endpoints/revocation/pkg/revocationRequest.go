package pkg

type RevocationRequest struct {
	Token      string `json:"token"`
	Recursive  bool   `json:"recursive"`
	OIDCIssuer string `json:"oidc_issuer"`
}
