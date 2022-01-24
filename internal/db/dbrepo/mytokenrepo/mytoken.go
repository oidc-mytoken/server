package mytokenrepo

import (
	"database/sql"
	"encoding/base64"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/db"
	eventService "github.com/oidc-mytoken/server/shared/mytoken/event"
	event "github.com/oidc-mytoken/server/shared/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/utils/cryptUtils"
	"github.com/oidc-mytoken/server/shared/utils/unixtime"
)

// MytokenEntry holds the information of a MytokenEntry as stored in the
// database
type MytokenEntry struct {
	ID                     mtid.MTID
	SeqNo                  uint64
	ParentID               mtid.MTID `db:"parent_id"`
	Token                  *mytoken.Mytoken
	rtID                   *uint64
	refreshToken           string
	encryptionKey          []byte
	rtEncrypted            string
	encryptionKeyEncrypted string
	Name                   string
	IP                     string `db:"ip_created"`
	networkData            api.ClientMetaData
	expiresAt              unixtime.UnixTime
}

// InitRefreshToken links a refresh token to this MytokenEntry
func (mte *MytokenEntry) InitRefreshToken(rt string) error {
	mte.refreshToken = rt
	mte.encryptionKey = cryptUtils.RandomBytes(32)
	tmp, err := cryptUtils.AESEncrypt(mte.refreshToken, mte.encryptionKey)
	if err != nil {
		return err
	}
	mte.rtEncrypted = tmp
	jwt, err := mte.Token.ToJWT()
	if err != nil {
		return err
	}
	tmp, err = cryptUtils.AES256Encrypt(base64.StdEncoding.EncodeToString(mte.encryptionKey), jwt)
	if err != nil {
		return err
	}
	mte.encryptionKeyEncrypted = tmp
	return nil
}

// SetRefreshToken updates the refresh token for this MytokenEntry
func (mte *MytokenEntry) SetRefreshToken(rtID uint64, key []byte) error {
	mte.encryptionKey = key
	jwt, err := mte.Token.ToJWT()
	if err != nil {
		return err
	}
	tmp, err := cryptUtils.AES256Encrypt(base64.StdEncoding.EncodeToString(key), jwt)
	if err != nil {
		return err
	}
	mte.encryptionKeyEncrypted = tmp
	mte.rtID = &rtID
	return nil
}

// NewMytokenEntry creates a new MytokenEntry
func NewMytokenEntry(mt *mytoken.Mytoken, name string, networkData api.ClientMetaData) *MytokenEntry {
	return &MytokenEntry{
		ID:          mt.ID,
		SeqNo:       mt.SeqNo,
		Token:       mt,
		Name:        name,
		IP:          networkData.IP,
		networkData: networkData,
		expiresAt:   mt.Restrictions.GetExpires(),
	}
}

// Root checks if this MytokenEntry is a root token
func (mte *MytokenEntry) Root() bool {
	return !mte.ParentID.HashValid()
}

// Store stores the MytokenEntry in the database
func (mte *MytokenEntry) Store(rlog log.Ext1FieldLogger, tx *sqlx.Tx, comment string) error {
	steStore := mytokenEntryStore{
		ID:        mte.ID,
		SeqNo:     mte.SeqNo,
		ParentID:  mte.ParentID,
		Name:      db.NewNullString(mte.Name),
		IP:        mte.IP,
		Iss:       mte.Token.OIDCIssuer,
		Sub:       mte.Token.OIDCSubject,
		ExpiresAt: db.NewNullTime(mte.expiresAt.Time()),
	}
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			if mte.rtID == nil {
				if _, err := tx.Exec(`CALL CryptStoreRT_Insert(?,@ID)`, mte.rtEncrypted); err != nil {
					return errors.WithStack(err)
				}
				var rtID uint64
				if err := tx.Get(&rtID, `SELECT @ID`); err != nil {
					return errors.WithStack(err)
				}
				mte.rtID = &rtID
			}
			steStore.RefreshTokenID = *mte.rtID
			if err := steStore.Store(rlog, tx); err != nil {
				return err
			}
			if err := storeEncryptionKey(tx, mte.encryptionKeyEncrypted, steStore.RefreshTokenID, mte.ID); err != nil {
				return err
			}
			return eventService.LogEvent(
				rlog, tx, eventService.MTEvent{
					Event: event.FromNumber(event.MTCreated, comment),
					MTID:  mte.ID,
				}, mte.networkData,
			)
		},
	)
}

func storeEncryptionKey(tx *sqlx.Tx, key string, rtID uint64, myid mtid.MTID) error {
	_, err := tx.Exec(`CALL EncryptionKeysRT_Insert(?,?,?)`, key, rtID, myid)
	return errors.WithStack(err)
}

type mytokenEntryStore struct {
	ID             mtid.MTID
	SeqNo          uint64
	ParentID       mtid.MTID `db:"parent_id"`
	RootID         mtid.MTID `db:"root_id"`
	RefreshTokenID uint64    `db:"rt_id"`
	Name           db.NullString
	IP             string `db:"ip_created"`
	Iss            string
	Sub            string
	ExpiresAt      sql.NullTime
}

// Store stores the mytokenEntryStore in the database; if this is the first token for this user, the user is also added
// to the db
func (e *mytokenEntryStore) Store(rlog log.Ext1FieldLogger, tx *sqlx.Tx) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(
				`CALL MTokens_Insert(?,?,?,?,?,?,?,?,?)`,
				e.Sub, e.Iss, e.ID, e.SeqNo, e.ParentID, e.RefreshTokenID, e.Name, e.IP, e.ExpiresAt,
			)
			return errors.WithStack(err)
		},
	)
}
