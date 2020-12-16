package pkg

// CreateTransferCodeRequest is a request to create a new transfer code from an existing super token
type CreateTransferCodeRequest struct {
	SuperToken string `json:"super_token"`
}
