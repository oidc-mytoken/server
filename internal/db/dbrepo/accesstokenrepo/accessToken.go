package accesstokenrepo

import (
	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/utils/cryptUtils"
)

// AccessToken holds database information about an access token
type AccessToken struct {
	Token   string
	IP      string
	Comment string
	Mytoken *mytoken.Mytoken

	Scopes    []string
	Audiences []string
}

type accessToken struct {
	Token   string
	IP      string `db:"ip_created"`
	Comment db.NullString
	STID    mtid.MTID `db:"ST_id"`
}

func (t *AccessToken) toDBObject() (*accessToken, error) {
	stJWT, err := t.Mytoken.ToJWT()
	if err != nil {
		return nil, err
	}
	token, err := cryptUtils.AES256Encrypt(t.Token, stJWT)
	if err != nil {
		return nil, err
	}
	return &accessToken{
		Token:   token,
		IP:      t.IP,
		Comment: db.NewNullString(t.Comment),
		STID:    t.Mytoken.ID,
	}, nil
}

func (t *AccessToken) getDBAttributes(tx *sqlx.Tx, atID uint64) (attrs []accessTokenAttribute, err error) {
	var scopeAttrID uint64
	var audAttrID uint64
	if err = tx.QueryRow(`SELECT id FROM Attributes WHERE attribute=?`, model.AttrScope).Scan(&scopeAttrID); err != nil {
		return
	}
	if err = tx.QueryRow(`SELECT id FROM Attributes WHERE attribute=?`, model.AttrAud).Scan(&audAttrID); err != nil {
		return
	}
	for _, s := range t.Scopes {
		attrs = append(attrs, accessTokenAttribute{
			ATID:   atID,
			AttrID: scopeAttrID,
			Attr:   s,
		})
	}
	for _, a := range t.Audiences {
		attrs = append(attrs, accessTokenAttribute{
			ATID:   atID,
			AttrID: audAttrID,
			Attr:   a,
		})
	}
	return
}

// Store stores the AccessToken in the database as well as the relevant attributes
func (t *AccessToken) Store(tx *sqlx.Tx) error {
	store, err := t.toDBObject()
	if err != nil {
		return err
	}
	storeFnc := func(tx *sqlx.Tx) error {
		res, err := tx.NamedExec(`INSERT INTO AccessTokens (token, ip_created, comment, ST_id) VALUES (:token, :ip_created, :comment, :ST_id)`, store)
		if err != nil {
			return err
		}
		if len(t.Scopes) > 0 || len(t.Audiences) > 0 {
			atID, err := res.LastInsertId()
			if err != nil {
				return err
			}
			attrs, err := t.getDBAttributes(tx, uint64(atID))
			if err != nil {
				return err
			}
			if _, err = tx.NamedExec(`INSERT INTO AT_Attributes (AT_id, attribute_id, attribute) VALUES (:AT_id, :attribute_id, :attribute)`, attrs); err != nil {
				return err
			}
		}
		return nil
	}
	return db.RunWithinTransaction(tx, storeFnc)
}
