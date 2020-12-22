package token

import (
	"encoding/json"
	"fmt"

	"github.com/zachmann/mytoken/internal/db/dbrepo/supertokenrepo/shorttokenrepo"
	"github.com/zachmann/mytoken/internal/utils"
)

// Token is a type used for tokens passed in http requests; these can be normal SuperTokens or a short token. This type will unmarshal always into a jwt of the (long) SuperToken
type Token string

// UnmarshalJSON implements the json.Unmarshaler interface
func (t *Token) UnmarshalJSON(data []byte) error {
	var token string
	if err := json.Unmarshal(data, &token); err != nil {
		return err
	}
	if utils.IsJWT(token) {
		*t = Token(token)
		return nil
	}
	shortToken, valid, err := shorttokenrepo.ParseShortToken(nil, token)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("token not valid")
	}
	token, err = shortToken.JWT()
	*t = Token(token)
	return err
}
