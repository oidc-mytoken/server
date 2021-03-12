package supertokenrepo

import (
	"encoding/base64"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/shared/supertoken/event"
	event "github.com/oidc-mytoken/server/shared/supertoken/event/pkg"
	supertoken "github.com/oidc-mytoken/server/shared/supertoken/pkg"
	"github.com/oidc-mytoken/server/shared/supertoken/pkg/stid"
	"github.com/oidc-mytoken/server/shared/utils/cryptUtils"
)

// SuperTokenEntry holds the information of a SuperTokenEntry as stored in the
// database
type SuperTokenEntry struct {
	ID                     stid.STID
	ParentID               stid.STID `db:"parent_id"`
	RootID                 stid.STID `db:"root_id"`
	Token                  *supertoken.SuperToken
	rtID                   *uint64
	refreshToken           string
	encryptionKey          []byte
	rtEncrypted            string
	encryptionKeyEncrypted string
	Name                   string
	IP                     string `db:"ip_created"`
	networkData            model.ClientMetaData
}

func (ste *SuperTokenEntry) InitRefreshToken(rt string) error {
	ste.refreshToken = rt
	ste.encryptionKey = cryptUtils.RandomBytes(32)
	tmp, err := cryptUtils.AESEncrypt(ste.refreshToken, ste.encryptionKey)
	if err != nil {
		return err
	}
	ste.rtEncrypted = tmp
	jwt, err := ste.Token.ToJWT()
	if err != nil {
		return err
	}
	tmp, err = cryptUtils.AES256Encrypt(base64.StdEncoding.EncodeToString(ste.encryptionKey), jwt)
	if err != nil {
		return err
	}
	ste.encryptionKeyEncrypted = tmp
	return nil
}

func (ste *SuperTokenEntry) SetRefreshToken(rtID uint64, key []byte) error {
	ste.encryptionKey = key
	jwt, err := ste.Token.ToJWT()
	if err != nil {
		return err
	}
	tmp, err := cryptUtils.AES256Encrypt(base64.StdEncoding.EncodeToString(key), jwt)
	if err != nil {
		return err
	}
	ste.encryptionKeyEncrypted = tmp
	ste.rtID = &rtID
	return nil
}

// NewSuperTokenEntry creates a new SuperTokenEntry
func NewSuperTokenEntry(st *supertoken.SuperToken, name string, networkData model.ClientMetaData) *SuperTokenEntry {
	return &SuperTokenEntry{
		ID:          st.ID,
		Token:       st,
		Name:        name,
		IP:          networkData.IP,
		networkData: networkData,
	}
}

// Root checks if this SuperTokenEntry is a root token
func (ste *SuperTokenEntry) Root() bool {
	return !ste.RootID.HashValid()
}

// Store stores the SuperTokenEntry in the database
func (ste *SuperTokenEntry) Store(tx *sqlx.Tx, comment string) error {
	steStore := superTokenEntryStore{
		ID:       ste.ID,
		ParentID: ste.ParentID,
		RootID:   ste.RootID,
		Name:     db.NewNullString(ste.Name),
		IP:       ste.IP,
		Iss:      ste.Token.OIDCIssuer,
		Sub:      ste.Token.OIDCSubject,
	}
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		if ste.rtID == nil {
			if _, err := tx.Exec(`INSERT INTO RefreshTokens  (rt)  VALUES(?)`, ste.rtEncrypted); err != nil {
				return err
			}
			var rtID uint64
			if err := tx.Get(&rtID, `SELECT LAST_INSERT_ID()`); err != nil {
				return err
			}
			ste.rtID = &rtID
		}
		steStore.RefreshTokenID = *ste.rtID
		if err := steStore.Store(tx); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT IGNORE INTO EncryptionKeys  (rt_id, MT_id, encryption_key)  VALUES(?,?,?)`, steStore.RefreshTokenID, ste.ID, ste.encryptionKeyEncrypted); err != nil {
			return err
		}
		return eventService.LogEvent(tx, event.FromNumber(event.STEventCreated, comment), ste.ID, ste.networkData)
	})
}

type superTokenEntryStore struct {
	ID             stid.STID
	ParentID       stid.STID `db:"parent_id"`
	RootID         stid.STID `db:"root_id"`
	RefreshTokenID uint64    `db:"rt_id"`
	Name           db.NullString
	IP             string `db:"ip_created"`
	Iss            string
	Sub            string
}

// Store stores the superTokenEntryStore in the database; if this is the first token for this user, the user is also added to the db
func (e *superTokenEntryStore) Store(tx *sqlx.Tx) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		stmt, err := tx.PrepareNamed(`INSERT INTO SuperTokens (id, parent_id, root_id, rt_id, name, ip_created, user_id) VALUES(:id, :parent_id, :root_id, :rt_id, :name, :ip_created, (SELECT id FROM Users WHERE iss=:iss AND sub=:sub))`)
		if err != nil {
			return err
		}
		txStmt := tx.NamedStmt(stmt)
		if _, err = txStmt.Exec(e); err != nil {
			var mysqlError *mysql.MySQLError
			if errors.As(err, &mysqlError) && mysqlError.Number == 1048 {
				_, err = tx.NamedExec(`INSERT INTO Users (sub, iss) VALUES(:sub, :iss)`, e)
				if err != nil {
					return err
				}
				_, err = txStmt.Exec(e)
				return err
			}
			log.WithError(err).Error()
			return err
		}
		return nil
	})
}
