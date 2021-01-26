package supertoken

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/supertokenrepo/transfercoderepo"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/super/pkg"
	"github.com/oidc-mytoken/server/internal/jws"
	serverModel "github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/supertoken/capabilities"
	eventService "github.com/oidc-mytoken/server/internal/supertoken/event"
	event "github.com/oidc-mytoken/server/internal/supertoken/event/pkg"
	"github.com/oidc-mytoken/server/internal/supertoken/restrictions"
	"github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/utils/issuerUtils"
)

// SuperToken is a mytoken SuperToken
type SuperToken struct {
	Issuer               string                    `json:"iss"`
	Subject              string                    `json:"sub"`
	ExpiresAt            int64                     `json:"exp,omitempty"`
	NotBefore            int64                     `json:"nbf"`
	IssuedAt             int64                     `json:"iat"`
	ID                   uuid.UUID                 `json:"jti"`
	Audience             string                    `json:"aud"`
	OIDCSubject          string                    `json:"oidc_sub"`
	OIDCIssuer           string                    `json:"oidc_iss"`
	Restrictions         restrictions.Restrictions `json:"restrictions,omitempty"`
	Capabilities         capabilities.Capabilities `json:"capabilities"`
	SubtokenCapabilities capabilities.Capabilities `json:"subtoken_capabilities,omitempty"`
	jwt                  string
}

func (st *SuperToken) verifyID() bool {
	if len(st.ID.String()) == 0 {
		return false
	}
	return true
}

func (st *SuperToken) verifySubject() bool {
	if len(st.Subject) == 0 {
		return false
	}
	if st.Subject != issuerUtils.CombineSubIss(st.OIDCSubject, st.OIDCIssuer) {
		return false
	}
	return true
}

// VerifyCapabilities verifies that this super token has the required capabilities
func (st *SuperToken) VerifyCapabilities(required ...capabilities.Capability) bool {
	if st.Capabilities == nil || len(st.Capabilities) == 0 {
		return false
	}
	for _, c := range required {
		if !st.Capabilities.Has(c) {
			return false
		}
	}
	return true
}

// NewSuperToken creates a new SuperToken
func NewSuperToken(oidcSub, oidcIss string, r restrictions.Restrictions, c, sc capabilities.Capabilities) *SuperToken {
	now := time.Now().Unix()
	st := &SuperToken{
		ID:                   uuid.NewV4(),
		IssuedAt:             now,
		NotBefore:            now,
		Issuer:               config.Get().IssuerURL,
		Subject:              issuerUtils.CombineSubIss(oidcSub, oidcIss),
		Audience:             config.Get().IssuerURL,
		OIDCIssuer:           oidcIss,
		OIDCSubject:          oidcSub,
		Capabilities:         c,
		SubtokenCapabilities: sc,
	}
	if len(r) > 0 {
		st.Restrictions = r
		exp := r.GetExpires()
		if exp != 0 {
			st.ExpiresAt = exp
		}
		nbf := r.GetNotBefore()
		if nbf != 0 && nbf > now {
			st.NotBefore = nbf
		}
	}
	return st
}

// ExpiresIn returns the amount of seconds in which this token expires
func (st *SuperToken) ExpiresIn() uint64 {
	now := time.Now().Unix()
	expAt := st.ExpiresAt
	if expAt > 0 && expAt > now {
		return uint64(expAt - now)
	}
	return 0
}

// Valid checks if this SuperToken is valid
func (st *SuperToken) Valid() error {
	standardClaims := jwt.StandardClaims{
		Audience:  st.Audience,
		ExpiresAt: st.ExpiresAt,
		Id:        st.ID.String(),
		IssuedAt:  st.IssuedAt,
		Issuer:    st.Issuer,
		NotBefore: st.NotBefore,
		Subject:   st.Subject,
	}
	if err := standardClaims.Valid(); err != nil {
		return err
	}
	if ok := standardClaims.VerifyIssuer(config.Get().IssuerURL, true); !ok {
		return fmt.Errorf("invalid issuer")
	}
	if ok := standardClaims.VerifyAudience(config.Get().IssuerURL, true); !ok {
		return fmt.Errorf("invalid Audience")
	}
	if ok := st.verifyID(); !ok {
		return fmt.Errorf("invalid id")
	}
	if ok := st.verifySubject(); !ok {
		return fmt.Errorf("invalid subject")
	}
	return nil
}

// ToSuperTokenResponse returns a SuperTokenResponse for this token. It requires that jwt is set or that the jwt is passed as argument; if not passed as argument toJWT must have been called earlier on this token to set jwt. This is always the case, if the token has been stored.
func (st *SuperToken) toSuperTokenResponse(jwt string) response.SuperTokenResponse {
	res := st.toTokenResponse()
	res.SuperToken = jwt
	res.SuperTokenType = model.ResponseTypeToken
	return res
}

func (st *SuperToken) toShortSuperTokenResponse(jwt string) (response.SuperTokenResponse, error) {
	shortToken, err := transfercoderepo.NewShortToken(jwt)
	if err != nil {
		return response.SuperTokenResponse{}, err
	}
	if err = shortToken.Store(nil); err != nil {
		return response.SuperTokenResponse{}, err
	}
	res := st.toTokenResponse()
	res.SuperToken = shortToken.String()
	res.SuperTokenType = model.ResponseTypeShortToken
	return res, nil
}

func (st *SuperToken) toTokenResponse() response.SuperTokenResponse {
	return response.SuperTokenResponse{
		ExpiresIn:            st.ExpiresIn(),
		Restrictions:         st.Restrictions,
		Capabilities:         st.Capabilities,
		SubtokenCapabilities: st.SubtokenCapabilities,
	}
}

// CreateTransferCode creates a transfer code for the passed super token
func CreateTransferCode(stid uuid.UUID, jwt string, newST bool, responseType model.ResponseType, clientMetaData serverModel.ClientMetaData) (string, uint64, error) {
	transferCode, err := transfercoderepo.NewTransferCode(jwt, newST, responseType)
	if err != nil {
		return "", 0, err
	}
	err = db.Transact(func(tx *sqlx.Tx) error {
		if err = transferCode.Store(tx); err != nil {
			return err
		}
		return eventService.LogEvent(tx, &event.Event{
			Type:    event.STEventTransferCodeCreated,
			Comment: fmt.Sprintf("token type: %s", responseType.String()),
		}, stid, clientMetaData)
	})
	expiresIn := uint64(config.Get().Features.Polling.PollingCodeExpiresAfter)
	return transferCode.String(), expiresIn, err
}

// ToTokenResponse creates a SuperTokenResponse for this SuperToken according to the passed model.ResponseType
func (st *SuperToken) ToTokenResponse(responseType model.ResponseType, networkData serverModel.ClientMetaData, jwt string) (response.SuperTokenResponse, error) {
	if len(jwt) == 0 {
		jwt = st.jwt
	}
	switch responseType {
	case model.ResponseTypeShortToken:
		if config.Get().Features.ShortTokens.Enabled {
			return st.toShortSuperTokenResponse(jwt)
		}
	case model.ResponseTypeTransferCode:
		transferCode, expiresIn, err := CreateTransferCode(st.ID, jwt, true, model.ResponseTypeToken, networkData)
		res := st.toTokenResponse()
		res.TransferCode = transferCode
		res.SuperTokenType = model.ResponseTypeTransferCode
		res.ExpiresIn = expiresIn
		return res, err
	}
	return st.toSuperTokenResponse(jwt), nil
}

// ToJWT returns the SuperToken as JWT
func (st *SuperToken) ToJWT() (string, error) {
	if st.jwt != "" {
		return st.jwt, nil
	}
	var err error
	st.jwt, err = jwt.NewWithClaims(jwt.GetSigningMethod(config.Get().Signing.Alg), st).SignedString(jws.GetPrivateKey())
	return st.jwt, err
}

// Value implements the driver.Valuer interface.
func (st *SuperToken) Value() (driver.Value, error) {
	return st.ToJWT()
}

// Scan implements the sql.Scanner interface.
func (st *SuperToken) Scan(src interface{}) error {
	tmp, err := ParseJWT(src.(string))
	if err != nil {
		return err
	}
	*st = *tmp
	return nil
}

// ParseJWT parses a token string into a SuperToken
func ParseJWT(token string) (*SuperToken, error) {
	tok, err := jwt.ParseWithClaims(token, &SuperToken{}, func(t *jwt.Token) (interface{}, error) {
		return jws.GetPublicKey(), nil
	})
	if err != nil {
		return nil, err
	}

	if st, ok := tok.Claims.(*SuperToken); ok && tok.Valid {
		return st, nil
	}
	return nil, fmt.Errorf("token not valid")
}
