package supertoken

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/db"
	"github.com/zachmann/mytoken/internal/jws"
	"github.com/zachmann/mytoken/internal/supertoken/capabilities"
	eventService "github.com/zachmann/mytoken/internal/supertoken/event"
	event "github.com/zachmann/mytoken/internal/supertoken/event/pkg"
	"github.com/zachmann/mytoken/internal/supertoken/restrictions"
	"github.com/zachmann/mytoken/internal/utils"
)

// SuperTokenEntry holds the information of a SuperTokenEntry as stored in the
// database
type SuperTokenEntry struct {
	ID           uuid.UUID
	ParentID     string `db:"parent_id"`
	RootID       string `db:"root_id"`
	Revoked      bool
	Token        *SuperToken
	RefreshToken string `db:"refresh_token"`
	Name         string
	CreatedAt    time.Time `db:"created_at"`
	IP           string    `db:"ip_created"`
}

type stEntryDBStoreObject struct {
	ID           uuid.UUID
	ParentID     sql.NullString `db:"parent_id"`
	RootID       sql.NullString `db:"root_id"`
	Revoked      bool
	Token        *SuperToken
	RefreshToken string `db:"refresh_token"`
	Name         string
	IP           string `db:"ip_created"`
	Iss          string
	Sub          string
}

// SuperToken is a mytoken SuperToken
type SuperToken struct {
	Issuer       string                    `json:"iss"`
	Subject      string                    `json:"sub"`
	ExpiresAt    int64                     `json:"exp,omitempty"`
	NotBefore    int64                     `json:"nbf"`
	IssuedAt     int64                     `json:"iat"`
	ID           uuid.UUID                 `json:"jti"`
	Audience     string                    `json:"aud"`
	OIDCSubject  string                    `json:"oidc_sub"`
	OIDCIssuer   string                    `json:"oidc_iss"`
	Restrictions restrictions.Restrictions `json:"restrictions,omitempty"`
	Capabilities capabilities.Capabilities `json:"capabilities"`
}

// SuperTokenEntryTree is a tree of SuperTokenEntry
type SuperTokenEntryTree struct {
	Token    SuperTokenEntry
	Children []SuperTokenEntryTree
}

func (t *SuperTokenEntryTree) print(level int) {
	for i := 0; i < 2*level; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("%s\n", t.Token.ID)
	for _, child := range t.Children {
		child.print(level + 1)
	}
}

// NewSuperToken creates a new SuperToken
func NewSuperToken(oidcSub, oidcIss string, r restrictions.Restrictions, c capabilities.Capabilities) *SuperToken {
	now := time.Now().Unix()
	st := &SuperToken{
		ID:           uuid.NewV4(),
		IssuedAt:     now,
		NotBefore:    now,
		Issuer:       config.Get().IssuerURL,
		Subject:      utils.CombineSubIss(oidcSub, oidcIss),
		Audience:     config.Get().IssuerURL,
		OIDCIssuer:   oidcIss,
		OIDCSubject:  oidcSub,
		Capabilities: c,
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

func newSuperTokenEntry(name, oidcSub, oidcIss string, r restrictions.Restrictions, c capabilities.Capabilities) *SuperTokenEntry {
	//TODO
	ip := "192.168.0.31"

	st := NewSuperToken(oidcSub, oidcIss, r, c)
	return &SuperTokenEntry{
		ID:    st.ID,
		Token: st,
		Name:  name,
		IP:    ip,
	}
}

func NewSuperTokenEntryFromSuperToken(name string, parent SuperTokenEntry, r restrictions.Restrictions, c capabilities.Capabilities) (*SuperTokenEntry, error) {
	newRestrictions := restrictions.Tighten(parent.Token.Restrictions, r)
	newCapabilities := capabilities.Tighten(parent.Token.Capabilities, c)
	ste := newSuperTokenEntry(name, parent.Token.OIDCSubject, parent.Token.OIDCIssuer, newRestrictions, newCapabilities)
	ste.ParentID = parent.ID.String()
	rootID := parent.ID.String()
	if !parent.Root() {
		rootID = parent.RootID
	}
	ste.RootID = rootID
	err := ste.Store("Used grant_type super_token")
	if err != nil {
		return nil, err
	}
	return ste, nil
}

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
		return fmt.Errorf("Invalid issuer")
	}
	if ok := standardClaims.VerifyAudience(config.Get().IssuerURL, true); !ok {
		return fmt.Errorf("Invalid Audience")
	}

	//TODO
	return nil
}

// ToJWT returns the SuperToken as JWT
func (st *SuperToken) ToJWT() (string, error) {
	return jwt.NewWithClaims(jwt.GetSigningMethod(config.Get().TokenSigningAlg), st).SignedString(jws.GetPrivateKey())
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
	return nil, fmt.Errorf("Propably token not valid")
}

func (ste *SuperTokenEntry) Root() bool {
	if ste.RootID == "" {
		return true
	}
	return false
}

func (ste *SuperTokenEntry) Store(comment string) error {
	steStore := stEntryDBStoreObject{
		ID:           ste.ID,
		ParentID:     db.NewNullString(ste.ParentID),
		RootID:       db.NewNullString(ste.RootID),
		Revoked:      ste.Revoked,
		Token:        ste.Token,
		RefreshToken: ste.RefreshToken,
		Name:         ste.Name,
		IP:           ste.IP,
		Iss:          ste.Token.OIDCIssuer,
		Sub:          ste.Token.OIDCSubject,
	}

	_, err := db.DB().NamedExec(`INSERT INTO SuperTokens (id, parent_id, root_id, revoked, token, refresh_token, name, ip_created, user_id) VALUES(:id, :parent_id, :root_id, :revoked, :token, :refresh_token, :name, :ip_created, (SELECT id FROM Users WHERE iss=:iss AND sub=:sub))`, steStore)
	if err != nil {
		return err
	}
	return eventService.LogEvent(*event.FromNumber(event.STEventSTCreated, comment), ste.ID)

}
