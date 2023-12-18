package shorttokenrepo

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/utils/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/utils/cryptutils"
	"github.com/oidc-mytoken/server/internal/utils/hashutils"
)

// ProxyToken holds information for proxy tokens, i.e. tokens that proxy another token, e.g. a short token
type ProxyToken struct {
	id    string
	token string
	mtID  mtid.MTID

	encryptedJWT string
	decryptedJWT string
}

// NewProxyToken creates a new ProxyToken of the given length
func NewProxyToken(size int) *ProxyToken {
	token := utils.RandReadableAlphaString(size)
	return CreateProxyToken(token)
}

func CreateProxyToken(token string) *ProxyToken {
	id := hashutils.SHA512Str([]byte(token))
	return &ProxyToken{
		id:    id,
		token: token,
	}
}

// ParseProxyToken parses the proxy token string into a proxyToken
func ParseProxyToken(token string) *ProxyToken {
	var id string
	if token != "" {
		id = hashutils.SHA512Str([]byte(token))
	}
	return &ProxyToken{
		id:    id,
		token: token,
	}
}

func (pt ProxyToken) String() string {
	return pt.Token()
}

// Token returns the token of this ProxyToken
func (pt ProxyToken) Token() string {
	return pt.token
}

// MTID returns the mtid.MTID of this ProxyToken
func (pt ProxyToken) MTID() mtid.MTID {
	return pt.mtID
}

// ID returns the id of this token
func (pt *ProxyToken) ID() string {
	if pt.id == "" {
		pt.id = hashutils.SHA512Str([]byte(pt.token))
	}
	return pt.id
}

// SetJWT sets the jwt for this proxyToken
func (pt *ProxyToken) SetJWT(jwt string, mID mtid.MTID) (err error) {
	pt.mtID = mID
	pt.decryptedJWT = jwt
	pt.encryptedJWT, err = cryptutils.AES256Encrypt(jwt, pt.token)
	return
}

// JWT returns the decrypted jwt that is linked to this proxyToken
func (pt *ProxyToken) JWT(rlog log.Ext1FieldLogger, tx *sqlx.Tx) (jwt string, valid bool, err error) {
	jwt = pt.decryptedJWT
	if jwt != "" {
		valid = true
		return
	}
	if pt.encryptedJWT == "" {
		if err = db.RunWithinTransaction(
			rlog, tx, func(tx *sqlx.Tx) error {
				var res struct {
					JWT  string    `db:"jwt"`
					MTID mtid.MTID `db:"MT_id"`
				}
				if err = tx.Get(&res, `CALL ProxyTokens_GetMT(?)`, pt.id); err != nil {
					return errors.WithStack(err)
				}
				pt.encryptedJWT = res.JWT
				pt.mtID = res.MTID
				return nil
			},
		); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = nil
				return
			}
			return
		}
	}
	valid = true
	if pt.encryptedJWT == "" {
		return
	}
	jwt, err = cryptutils.AES256Decrypt(pt.encryptedJWT, pt.token)
	pt.decryptedJWT = jwt
	return
}

// Store stores the proxyToken
func (pt ProxyToken) Store(rlog log.Ext1FieldLogger, tx *sqlx.Tx) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL ProxyTokens_Insert(?,?,?)`, pt.id, pt.encryptedJWT, pt.mtID)
			return errors.WithStack(err)
		},
	)
}

// Update updates the jwt of the proxyToken
func (pt ProxyToken) Update(rlog log.Ext1FieldLogger, tx *sqlx.Tx) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL ProxyTokens_Update(?,?,?)`, pt.ID(), pt.encryptedJWT, pt.mtID)
			return errors.WithStack(err)
		},
	)
}

// Delete deletes the proxyToken from the database, it does not delete the linked Mytoken, the jwt should have been
// retrieved earlier and the Mytoken if desired be revoked separately
func (pt ProxyToken) Delete(rlog log.Ext1FieldLogger, tx *sqlx.Tx) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL ProxyTokens_Delete(?)`, pt.id)
			return errors.WithStack(err)
		},
	)
}
