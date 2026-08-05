package accesstokenrepo

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/encryptionkeyrepo"
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

	ExpiresAt time.Time
	TokenType string
}

type accessToken struct {
	Token     string
	ExpiresAt sql.NullTime  `db:"expires_at"`
	TokenType db.NullString `db:"token_type"`
	IP        string        `db:"ip_created"`
	Comment   db.NullString
	MTID      mtid.MTID `db:"MT_id"`
}

// Store stores the AccessToken in the database as well as the relevant attributes. The access token is encrypted with
// the mytoken's encryption key so that it stays decryptable if the mytoken is rotated.
func (t *AccessToken) Store(rlog log.Ext1FieldLogger, tx *sqlx.Tx) error {
	stJWT, err := t.Mytoken.ToJWT()
	if err != nil {
		return err
	}
	key, _, err := encryptionkeyrepo.GetEncryptionKey(rlog, tx, t.Mytoken.ID, stJWT)
	if err != nil {
		return errors.WithStack(err)
	}
	encryptedAT, err := cryptutils.AESEncrypt(t.Token, key)
	if err != nil {
		return errors.WithStack(err)
	}
	store := &accessToken{
		Token:     encryptedAT,
		ExpiresAt: db.NewNullTime(t.ExpiresAt),
		TokenType: db.NewNullString(t.TokenType),
		IP:        t.IP,
		Comment:   db.NewNullString(t.Comment),
		MTID:      t.Mytoken.ID,
	}
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			var atID uint64
			err = tx.Get(
				&atID, `CALL AT_Insert(?,?,?,?,?,?)`,
				store.Token, store.IP, store.Comment, store.MTID,
				store.ExpiresAt, store.TokenType,
			)
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

// scopesJSON returns the passed list of scopes as a json array string; the list is used for matching an access token
// against the scopes/audiences stored as access token attributes
func scopesJSON(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	data, _ := json.Marshal(values)
	return string(data)
}

// CachedAccessToken is a reusable access token obtained from the cache
type CachedAccessToken struct {
	Token     string
	TokenType string
	Created   time.Time
	ExpiresAt time.Time
}

type atCachedDB struct {
	Crypt     string    `db:"crypt"`
	Created   time.Time `db:"created"`
	ExpiresAt time.Time `db:"expires_at"`
	TokenType string    `db:"token_type"`
}

// GetCachedAT checks the database for a reusable access token matching the passed mytoken, scopes and audiences; it
// returns nil if no matching access token is stored
func GetCachedAT(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mytokenID mtid.MTID, jwt string, scopes, auds []string,
) (*CachedAccessToken, error) {
	var res atCachedDB
	found, err := db.ParseError(
		db.RunWithinTransaction(
			rlog, tx, func(tx *sqlx.Tx) error {
				return errors.WithStack(
					tx.Get(
						&res, `CALL AT_GetCached(?,?,?)`, mytokenID, scopesJSON(scopes), scopesJSON(auds),
					),
				)
			},
		),
	)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	key, _, err := encryptionkeyrepo.GetEncryptionKey(rlog, tx, mytokenID, jwt)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	plainAT, err := cryptutils.AESDecrypt(res.Crypt, key)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &CachedAccessToken{
		Token:     plainAT,
		TokenType: res.TokenType,
		Created:   res.Created,
		ExpiresAt: res.ExpiresAt,
	}, nil
}
