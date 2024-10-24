package universalmytoken

import (
	"encoding/json"

	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/utils/jwtutils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/shorttokenrepo"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/utils"
)

// UniversalMytoken is a type used for Mytokens passed in http requests; these can be normal Mytoken or a short token.
// This type will always provide a jwt of the (long) Mytoken
type UniversalMytoken struct {
	JWT               string
	OriginalToken     string
	OriginalTokenType model.ResponseType
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (t *UniversalMytoken) UnmarshalText(data []byte) (err error) {
	s := string(data)
	*t, err = Parse(log.StandardLogger(), s)
	return errors.WithStack(err)
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
	if len(token) < api.MinShortTokenLen {
		return UniversalMytoken{}, errors.New("token not valid")
	}
	if jwtutils.IsJWT(token) {
		return UniversalMytoken{
			JWT:               token,
			OriginalToken:     token,
			OriginalTokenType: model.ResponseTypeToken,
		}, nil
	}
	shortToken := shorttokenrepo.ParseShortToken(token)
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
