package pkg

// CreateTransferCodeRequest is a request to create a new transfer code from an existing super token
type CreateTransferCodeRequest struct {
	SuperToken string `json:"super_token"` // we use string and not token.Token because the token can also be in the Auth Header and there it is a string
}
