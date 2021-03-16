package mytoken

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/jws"
	serverModel "github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/mytoken/capabilities"
	eventService "github.com/oidc-mytoken/server/shared/mytoken/event"
	event "github.com/oidc-mytoken/server/shared/mytoken/event/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
	"github.com/oidc-mytoken/server/shared/utils/issuerUtils"
	"github.com/oidc-mytoken/server/shared/utils/unixtime"
)

// Mytoken is a mytoken Mytoken
type Mytoken struct {
	Issuer               string                    `json:"iss"`
	Subject              string                    `json:"sub"`
	ExpiresAt            unixtime.UnixTime         `json:"exp,omitempty"`
	NotBefore            unixtime.UnixTime         `json:"nbf"`
	IssuedAt             unixtime.UnixTime         `json:"iat"`
	ID                   mtid.MTID                 `json:"jti"`
	Audience             string                    `json:"aud"`
	OIDCSubject          string                    `json:"oidc_sub"`
	OIDCIssuer           string                    `json:"oidc_iss"`
	Restrictions         restrictions.Restrictions `json:"restrictions,omitempty"`
	Capabilities         capabilities.Capabilities `json:"capabilities"`
	SubtokenCapabilities capabilities.Capabilities `json:"subtoken_capabilities,omitempty"`
	jwt                  string
}

func (st *Mytoken) verifyID() bool {
	return st.ID.Valid()
}

func (st *Mytoken) verifySubject() bool {
	if st.Subject == "" {
		return false
	}
	if st.Subject != issuerUtils.CombineSubIss(st.OIDCSubject, st.OIDCIssuer) {
		return false
	}
	return true
}

// VerifyCapabilities verifies that this Mytoken has the required capabilities
func (st *Mytoken) VerifyCapabilities(required ...capabilities.Capability) bool {
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

// NewMytoken creates a new Mytoken
func NewMytoken(oidcSub, oidcIss string, r restrictions.Restrictions, c, sc capabilities.Capabilities) *Mytoken {
	now := unixtime.Now()
	st := &Mytoken{
		ID:                   mtid.New(),
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
func (st *Mytoken) ExpiresIn() uint64 {
	now := unixtime.Now()
	expAt := st.ExpiresAt
	if expAt > 0 && expAt > now {
		return uint64(expAt - now)
	}
	return 0
}

// Valid checks if this Mytoken is valid
func (st *Mytoken) Valid() error {
	standardClaims := jwt.StandardClaims{
		Audience:  st.Audience,
		ExpiresAt: int64(st.ExpiresAt),
		Id:        st.ID.String(),
		IssuedAt:  int64(st.IssuedAt),
		Issuer:    st.Issuer,
		NotBefore: int64(st.NotBefore),
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

// ToMytokenResponse returns a MytokenResponse for this token. It requires that jwt is set or that the jwt is passed as argument; if not passed as argument toJWT must have been called earlier on this token to set jwt. This is always the case, if the token has been stored.
func (st *Mytoken) toMytokenResponse(jwt string) response.MytokenResponse {
	res := st.toTokenResponse()
	res.Mytoken = jwt
	res.MytokenType = model.ResponseTypeToken
	return res
}

func (st *Mytoken) toShortMytokenResponse(jwt string) (response.MytokenResponse, error) {
	shortToken, err := transfercoderepo.NewShortToken(jwt, st.ID)
	if err != nil {
		return response.MytokenResponse{}, err
	}
	if err = shortToken.Store(nil); err != nil {
		return response.MytokenResponse{}, err
	}
	res := st.toTokenResponse()
	res.Mytoken = shortToken.String()
	res.MytokenType = model.ResponseTypeShortToken
	return res, nil
}

func (st *Mytoken) toTokenResponse() response.MytokenResponse {
	return response.MytokenResponse{
		ExpiresIn:            st.ExpiresIn(),
		Restrictions:         st.Restrictions,
		Capabilities:         st.Capabilities,
		SubtokenCapabilities: st.SubtokenCapabilities,
	}
}

// CreateTransferCode creates a transfer code for the passed mytoken id
func CreateTransferCode(stid mtid.MTID, jwt string, newST bool, responseType model.ResponseType, clientMetaData serverModel.ClientMetaData) (string, uint64, error) {
	transferCode, err := transfercoderepo.NewTransferCode(jwt, stid, newST, responseType)
	if err != nil {
		return "", 0, err
	}
	err = db.Transact(func(tx *sqlx.Tx) error {
		if err = transferCode.Store(tx); err != nil {
			return err
		}
		return eventService.LogEvent(tx, eventService.MTEvent{
			Event: event.FromNumber(event.MTEventTransferCodeCreated, fmt.Sprintf("token type: %s", responseType.String())),
			MTID:  stid,
		}, clientMetaData)
	})
	expiresIn := uint64(config.Get().Features.Polling.PollingCodeExpiresAfter)
	return transferCode.String(), expiresIn, err
}

// ToTokenResponse creates a MytokenResponse for this Mytoken according to the passed model.ResponseType
func (st *Mytoken) ToTokenResponse(responseType model.ResponseType, networkData serverModel.ClientMetaData, jwt string) (response.MytokenResponse, error) {
	if jwt == "" {
		jwt = st.jwt
	}
	switch responseType {
	case model.ResponseTypeShortToken:
		if config.Get().Features.ShortTokens.Enabled {
			return st.toShortMytokenResponse(jwt)
		}
	case model.ResponseTypeTransferCode:
		transferCode, expiresIn, err := CreateTransferCode(st.ID, jwt, true, model.ResponseTypeToken, networkData)
		res := st.toTokenResponse()
		res.TransferCode = transferCode
		res.MytokenType = model.ResponseTypeTransferCode
		res.ExpiresIn = expiresIn
		return res, err
	}
	return st.toMytokenResponse(jwt), nil
}

// ToJWT returns the Mytoken as JWT
func (st *Mytoken) ToJWT() (string, error) {
	if st.jwt != "" {
		return st.jwt, nil
	}
	var err error
	st.jwt, err = jwt.NewWithClaims(jwt.GetSigningMethod(config.Get().Signing.Alg), st).SignedString(jws.GetPrivateKey())
	return st.jwt, err
}

// ParseJWT parses a token string into a Mytoken
func ParseJWT(token string) (*Mytoken, error) {
	tok, err := jwt.ParseWithClaims(token, &Mytoken{}, func(t *jwt.Token) (interface{}, error) {
		return jws.GetPublicKey(), nil
	})
	if err != nil {
		return nil, err
	}

	if st, ok := tok.Claims.(*Mytoken); ok && tok.Valid {
		return st, nil
	}
	return nil, fmt.Errorf("token not valid")
}
