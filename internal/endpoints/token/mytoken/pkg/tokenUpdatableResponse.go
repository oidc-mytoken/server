package pkg

// TokenUpdatableResponse is an interface for responses that can have a MytokenResponse as a token update (after token
// rotation)
type TokenUpdatableResponse interface {
	SetTokenUpdate(response *MytokenResponse)
}
