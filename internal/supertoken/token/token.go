package token

import (
	"encoding/json"
	"fmt"

	"github.com/zachmann/mytoken/internal/db/dbrepo/supertokenrepo/transfercoderepo"
	"github.com/zachmann/mytoken/internal/utils"
)

// Token is a type used for tokens passed in http requests; these can be normal SuperTokens or a short token. This type will unmarshal always into a jwt of the (long) SuperToken
type Token string

// UnmarshalJSON implements the json.Unmarshaler interface
func (t *Token) UnmarshalJSON(data []byte) (err error) {
	var token string
	if err = json.Unmarshal(data, &token); err != nil {
		return
	}
	*t, err = GetLongSuperToken(token)
	return
}

// GetLongSuperToken returns the long / jwt of a super token; the passed token can be a jwt or a short token
func GetLongSuperToken(token string) (Token, error) {
	if utils.IsJWT(token) {
		return Token(token), nil
	}
	shortToken := transfercoderepo.ParseShortToken(token)
	token, valid, dbErr := shortToken.JWT(nil)
	var validErr error
	if !valid {
		validErr = fmt.Errorf("token not valid")
	}
	return Token(token), utils.ORErrors(dbErr, validErr)
}
