package accesstokenrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/utils/cryptutils"
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
	MTID    mtid.MTID `db:"MT_id"`
}

func (t *AccessToken) toDBObject() (*accessToken, error) {
	stJWT, err := t.Mytoken.ToJWT()
	if err != nil {
		return nil, err
	}
	token, err := cryptutils.AES256Encrypt(t.Token, stJWT)
	if err != nil {
		return nil, err
	}
	return &accessToken{
		Token:   token,
		IP:      t.IP,
		Comment: db.NewNullString(t.Comment),
		MTID:    t.Mytoken.ID,
	}, nil
}

// Store stores the AccessToken in the database as well as the relevant attributes
func (t *AccessToken) Store(rlog log.Ext1FieldLogger, tx *sqlx.Tx) error {
	store, err := t.toDBObject()
	if err != nil {
		return err
	}
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			var atID uint64
			err = tx.Get(&atID, `CALL AT_Insert(?,?,?,?)`, store.Token, store.IP, store.Comment, store.MTID)
			if err != nil {
				return errors.WithStack(err)
			}
			for _, s := range t.Scopes {
				if _, err = tx.Exec(`CALL ATAttribute_Insert(?,?,?)`, atID, s, model.AttrScope); err != nil {
					return errors.WithStack(err)
				}
			}
			for _, a := range t.Audiences {
				if _, err = tx.Exec(`CALL ATAttribute_Insert(?,?,?)`, atID, a, model.AttrAud); err != nil {
					return errors.WithStack(err)
				}
			}
			return nil
		},
	)
}
