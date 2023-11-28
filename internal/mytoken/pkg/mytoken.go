package mytoken

import (
	"encoding/json"
	"fmt"

	"github.com/golang-jwt/jwt"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/utils/unixtime"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/jws"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/utils"
)

// Mytoken is a mytoken Mytoken
type Mytoken struct {
	api.Mytoken
	ExpiresAt    unixtime.UnixTime         `json:"exp,omitempty"`
	NotBefore    unixtime.UnixTime         `json:"nbf"`
	IssuedAt     unixtime.UnixTime         `json:"iat"`
	AuthTime     unixtime.UnixTime         `json:"auth_time,omitempty"`
	ID           mtid.MTID                 `json:"jti"`
	Restrictions restrictions.Restrictions `json:"restrictions,omitempty"`
	Rotation     *api.Rotation             `json:"rotation,omitempty"`
	jwt          string
}

// ToUniversalMytoken returns a universalmytoken.UniversalMytoken for this Mytoken
func (mt *Mytoken) ToUniversalMytoken() universalmytoken.UniversalMytoken {
	j, _ := mt.ToJWT()
	return universalmytoken.UniversalMytoken{
		JWT:               j,
		OriginalToken:     j,
		OriginalTokenType: model.ResponseTypeToken,
	}
}

// Rotate returns a Mytoken and returns the new *Mytoken
func (mt Mytoken) Rotate() *Mytoken { // skipcq: CRT-P0003
	rotated := mt
	rotated.SeqNo++
	if rotated.Rotation.Lifetime > 0 {
		rotated.ExpiresAt = unixtime.InSeconds(int64(rotated.Rotation.Lifetime))
	}
	rotated.IssuedAt = unixtime.Now()
	rotated.NotBefore = rotated.IssuedAt
	rotated.jwt = ""
	return &rotated
}

func (mt *Mytoken) verifyID() bool {
	return mt.ID.Valid()
}

func (mt *Mytoken) verifySubject() bool {
	if mt.Subject == "" {
		return false
	}
	if mt.Subject != utils.CreateMytokenSubject(mt.OIDCSubject, mt.OIDCIssuer) {
		return false
	}
	return true
}

// VerifyCapabilities verifies that this Mytoken has the required capabilities
func (mt *Mytoken) VerifyCapabilities(required ...api.Capability) bool {
	if mt.Capabilities == nil || len(mt.Capabilities) == 0 {
		return false
	}
	for _, c := range required {
		if !mt.Capabilities.Has(c) {
			return false
		}
	}
	return true
}

// NewMytoken creates a new Mytoken
func NewMytoken(
	oidcSub, oidcIss, name string, r restrictions.Restrictions, c api.Capabilities, rot *api.Rotation,
	authTime unixtime.UnixTime,
) (*Mytoken, error) {
	now := unixtime.Now()
	id, err := mtid.New()
	if err != nil {
		return nil, err
	}
	mt := &Mytoken{
		Mytoken: api.Mytoken{
			Version:      api.TokenVer,
			TokenType:    api.TokenType,
			SeqNo:        1,
			Name:         name,
			Issuer:       config.Get().IssuerURL,
			Subject:      utils.CreateMytokenSubject(oidcSub, oidcIss),
			Audience:     config.Get().IssuerURL,
			OIDCIssuer:   oidcIss,
			OIDCSubject:  oidcSub,
			Capabilities: c,
		},
		ID:        id,
		IssuedAt:  now,
		NotBefore: now,
		AuthTime:  authTime,
		Rotation:  rot,
	}
	r.EnforceMaxLifetime(oidcIss)
	if len(r) > 0 {
		mt.Restrictions = r
		exp := r.GetExpires()
		if exp != 0 {
			mt.ExpiresAt = exp
		}
		nbf := r.GetNotBefore()
		if nbf != 0 && nbf > now {
			mt.NotBefore = nbf
		}
	}
	return mt, nil
}

// ExpiresIn returns the amount of seconds in which this token expires
func (mt *Mytoken) ExpiresIn() uint64 {
	now := unixtime.Now()
	expAt := mt.ExpiresAt
	if expAt > 0 && expAt > now {
		return uint64(expAt - now)
	}
	return 0
}

// Valid checks if this Mytoken is valid
func (mt *Mytoken) Valid() error {
	standardClaims := jwt.StandardClaims{
		Audience:  mt.Audience,
		ExpiresAt: int64(mt.ExpiresAt),
		Id:        mt.ID.String(),
		IssuedAt:  int64(mt.IssuedAt),
		Issuer:    mt.Issuer,
		NotBefore: int64(mt.NotBefore),
		Subject:   mt.Subject,
	}
	if err := errors.WithStack(standardClaims.Valid()); err != nil {
		return err
	}
	if ok := standardClaims.VerifyIssuer(config.Get().IssuerURL, true); !ok {
		return errors.New("invalid issuer")
	}
	if ok := standardClaims.VerifyAudience(config.Get().IssuerURL, true); !ok {
		return errors.New("invalid Audience")
	}
	if ok := mt.verifyID(); !ok {
		return errors.New("invalid id")
	}
	if ok := mt.verifySubject(); !ok {
		return errors.New("invalid subject")
	}
	return nil
}

// toMytokenResponse returns a pkg.MytokenResponse for this token. It requires that jwt is set or that the jwt is passed
// as argument; if not passed as argument toJWT must have been called earlier on this token to set jwt. This is always
// the case, if the token has been stored.
func (mt *Mytoken) toMytokenResponse(jwt string) response.MytokenResponse {
	res := mt.toTokenResponse()
	res.Mytoken = jwt
	res.MytokenType = model.ResponseTypeToken
	return res
}

func (mt *Mytoken) toShortMytokenResponse(rlog log.Ext1FieldLogger, jwt string) (response.MytokenResponse, error) {
	shortToken, err := transfercoderepo.NewShortToken(jwt, mt.ID)
	if err != nil {
		return response.MytokenResponse{}, err
	}
	if err = shortToken.Store(rlog, nil); err != nil {
		return response.MytokenResponse{}, err
	}
	res := mt.toTokenResponse()
	res.Mytoken = shortToken.String()
	res.MytokenType = model.ResponseTypeShortToken
	return res, nil
}

func (mt *Mytoken) toTokenResponse() response.MytokenResponse {
	return response.MytokenResponse{
		MytokenResponse: api.MytokenResponse{
			ExpiresIn:    mt.ExpiresIn(),
			Capabilities: mt.Capabilities,
			Rotation:     mt.Rotation,
			MOMID:        mt.ID.Hash(),
		},
		Restrictions: mt.Restrictions,
	}
}

// CreateTransferCode creates a transfer code for the passed mytoken id
func CreateTransferCode(
	rlog log.Ext1FieldLogger, myID mtid.MTID, jwt string, newMT bool, responseType model.ResponseType,
	clientMetaData api.ClientMetaData,
) (string, uint64, error) {
	transferCode, err := transfercoderepo.NewTransferCode(jwt, myID, newMT, responseType)
	if err != nil {
		return "", 0, err
	}
	err = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if err = transferCode.Store(rlog, tx); err != nil {
				return err
			}
			return eventService.LogEvent(
				rlog, tx, eventService.MTEvent{
					Event:   api.EventTransferCodeCreated,
					Comment: fmt.Sprintf("token type: %s", responseType.String()),
					MTID:    myID,
				}, clientMetaData,
			)
		},
	)
	expiresIn := uint64(config.Get().Features.Polling.PollingCodeExpiresAfter)
	return transferCode.String(), expiresIn, err
}

// ToTokenResponse creates a MytokenResponse for this Mytoken according to the passed model.ResponseType
func (mt *Mytoken) ToTokenResponse(
	rlog log.Ext1FieldLogger, responseType model.ResponseType, maxTokenLen int, networkData api.ClientMetaData,
	jwt string,
) (response.MytokenResponse, error) {
	if jwt == "" {
		jwt = mt.jwt
	}
	if maxTokenLen > 0 {
		if maxTokenLen >= len(jwt) {
			responseType = model.ResponseTypeToken
		} else if config.Get().Features.ShortTokens.Enabled && maxTokenLen >= config.Get().Features.ShortTokens.Len {
			responseType = model.ResponseTypeShortToken
		} else if config.Get().Features.TransferCodes.Enabled {
			responseType = model.ResponseTypeTransferCode
		} else {
			responseType = model.ResponseTypeToken
		}
	}
	switch responseType {
	case model.ResponseTypeShortToken:
		if config.Get().Features.ShortTokens.Enabled {
			return mt.toShortMytokenResponse(rlog, jwt)
		}
	case model.ResponseTypeTransferCode:
		transferCode, expiresIn, err := CreateTransferCode(rlog, mt.ID, jwt, true, model.ResponseTypeToken, networkData)
		res := mt.toTokenResponse()
		res.TransferCode = transferCode
		res.MytokenType = model.ResponseTypeTransferCode
		res.ExpiresIn = expiresIn
		res.MOMID = mt.ID.Hash()
		return res, err
	}
	return mt.toMytokenResponse(jwt), nil
}

// ToJWT returns the Mytoken as JWT
func (mt *Mytoken) ToJWT() (string, error) {
	if mt.jwt != "" {
		return mt.jwt, nil
	}
	var err error
	j := jwt.NewWithClaims(
		jwt.GetSigningMethod(config.Get().Signing.Mytoken.Alg.String()), mt,
	)
	j.Header["typ"] = "MT+JWT"
	mt.jwt, err = j.SignedString(jws.GetSigningKey(jws.KeyUsageMytokenSigning))
	return mt.jwt, errors.WithStack(err)
}

// ParseJWT parses a token string into a Mytoken
func ParseJWT(token string) (*Mytoken, error) {
	return parseJWT(token, false)
}

// ParseJWTWithoutClaimsValidation parses a token string into a Mytoken
func ParseJWTWithoutClaimsValidation(token string) (*Mytoken, error) {
	return parseJWT(token, true)
}

func parseJWT(token string, skipCalimsValidation bool) (*Mytoken, error) {
	parser := jwt.Parser{
		SkipClaimsValidation: skipCalimsValidation,
	}
	tok, err := parser.ParseWithClaims(
		token, &Mytoken{}, func(t *jwt.Token) (interface{}, error) {
			return jws.GetPublicKey(jws.KeyUsageMytokenSigning), nil
		},
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if mt, ok := tok.Claims.(*Mytoken); ok && tok.Valid {
		mt.jwt = token
		return mt, specialTokenHandling(mt)
	}
	return nil, errors.New("token not valid")
}

func (mt *Mytoken) DBMetadata() (meta mytokenrepohelper.MytokenDBMetadata, err error) {
	creator := func(i any) (db.NullString, error) {
		data := ""
		if i != nil {
			dataBytes, err := json.Marshal(i)
			if err != nil {
				return db.NullString{}, errors.WithStack(err)
			}
			data = string(dataBytes)
		}
		return db.NewNullString(data), nil
	}

	meta.Capabilities, err = creator(mt.Capabilities)
	if err != nil {
		return
	}
	meta.Rotation, err = creator(mt.Rotation)
	if err != nil {
		return
	}
	meta.Restrictions, err = creator(mt.Restrictions)
	if err != nil {
		return
	}
	return
}

func specialTokenHandling(mt *Mytoken) error {
	if mt.Version.Before(
		api.TokenVersion{
			Major: 0,
			Minor: 7,
		},
	) {
		_ = db.Transact(
			log.StandardLogger(), func(tx *sqlx.Tx) error {
				meta, err := mt.DBMetadata()
				if err != nil {
					return err
				}
				return mytokenrepohelper.SetMetadata(
					log.StandardLogger(), tx, mt.ID, meta,
				)
			},
		)
	}
	return nil
}
