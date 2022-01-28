package universalmytoken

import (
	"encoding/json"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	"github.com/oidc-mytoken/server/shared/model"
	"github.com/oidc-mytoken/server/shared/utils"
)

// UniversalMytoken is a type used for Mytokens passed in http requests; these can be normal Mytoken or a short token.
// This type will always provide a jwt of the (long) Mytoken
type UniversalMytoken struct {
	JWT               string
	OriginalToken     string
	OriginalTokenType model.ResponseType
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (t *UniversalMytoken) UnmarshalJSON(data []byte) (err error) {
	var token string
	if err = errors.WithStack(json.Unmarshal(data, &token)); err != nil {
		return
	}
	*t, err = Parse(log.StandardLogger(), token)
	return errors.WithStack(err)
}

// Parse parses a mytoken string (that can be a long or short mytoken) into an UniversalMytoken holding the JWT
func Parse(rlog log.Ext1FieldLogger, token string) (UniversalMytoken, error) {
	if token == "" {
		return UniversalMytoken{}, errors.New("token not valid")
	}
	if utils.IsJWT(token) {
		return UniversalMytoken{
			JWT:               token,
			OriginalToken:     token,
			OriginalTokenType: model.ResponseTypeToken,
		}, nil
	}
	shortToken := transfercoderepo.ParseShortToken(token)
	jwt, valid, dbErr := shortToken.JWT(rlog, nil)
	var validErr error
	if !valid {
		validErr = errors.New("token not valid")
	}
	return UniversalMytoken{
		JWT:               jwt,
		OriginalToken:     token,
		OriginalTokenType: model.ResponseTypeShortToken,
	}, utils.ORErrors(dbErr, validErr)
}
