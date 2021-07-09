package universalmytoken

import (
	"encoding/json"
	"fmt"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	"github.com/oidc-mytoken/server/shared/model"
	"github.com/oidc-mytoken/server/shared/utils"
)

// UniversalMytoken is a type used for Mytokens passed in http requests; these can be normal Mytoken or a short token. This type will always provide a jwt of the (long) Mytoken
type UniversalMytoken struct {
	JWT               string
	OriginalToken     string
	OriginalTokenType model.ResponseType
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (t *UniversalMytoken) UnmarshalJSON(data []byte) (err error) {
	var token string
	if err = json.Unmarshal(data, &token); err != nil {
		return
	}
	*t, err = Parse(token)
	return
}

// Parse parses a mytoken string (that can be a long or short mytoken) into an UniversalMytoken holding the JWT
func Parse(token string) (UniversalMytoken, error) {
	if token == "" {
		return UniversalMytoken{}, fmt.Errorf("token not valid")
	}
	if utils.IsJWT(token) {
		return UniversalMytoken{
			JWT:               token,
			OriginalToken:     token,
			OriginalTokenType: model.ResponseTypeToken,
		}, nil
	}
	shortToken := transfercoderepo.ParseShortToken(token)
	jwt, valid, dbErr := shortToken.JWT(nil)
	var validErr error
	if !valid {
		validErr = fmt.Errorf("token not valid")
	}
	return UniversalMytoken{
		JWT:               jwt,
		OriginalToken:     token,
		OriginalTokenType: model.ResponseTypeShortToken,
	}, utils.ORErrors(dbErr, validErr)
}
