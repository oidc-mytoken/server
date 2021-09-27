package mytokenrepo

import (
	"encoding/base64"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/db"
	eventService "github.com/oidc-mytoken/server/shared/mytoken/event"
	event "github.com/oidc-mytoken/server/shared/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/utils/cryptUtils"
)

// MytokenEntry holds the information of a MytokenEntry as stored in the
// database
type MytokenEntry struct {
	ID                     mtid.MTID
	SeqNo                  uint64
	ParentID               mtid.MTID `db:"parent_id"`
	RootID                 mtid.MTID `db:"root_id"`
	Token                  *mytoken.Mytoken
	rtID                   *uint64
	refreshToken           string
	encryptionKey          []byte
	rtEncrypted            string
	encryptionKeyEncrypted string
	Name                   string
	IP                     string `db:"ip_created"`
	networkData            api.ClientMetaData
}

// InitRefreshToken links a refresh token to this MytokenEntry
func (ste *MytokenEntry) InitRefreshToken(rt string) error {
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

// SetRefreshToken updates the refresh token for this MytokenEntry
func (ste *MytokenEntry) SetRefreshToken(rtID uint64, key []byte) error {
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

// NewMytokenEntry creates a new MytokenEntry
func NewMytokenEntry(mt *mytoken.Mytoken, name string, networkData api.ClientMetaData) *MytokenEntry {
	return &MytokenEntry{
		ID:          mt.ID,
		SeqNo:       mt.SeqNo,
		Token:       mt,
		Name:        name,
		IP:          networkData.IP,
		networkData: networkData,
	}
}

// Root checks if this MytokenEntry is a root token
func (ste *MytokenEntry) Root() bool {
	return !ste.RootID.HashValid()
}

// Store stores the MytokenEntry in the database
func (ste *MytokenEntry) Store(tx *sqlx.Tx, comment string) error {
	steStore := mytokenEntryStore{
		ID:       ste.ID,
		SeqNo:    ste.SeqNo,
		ParentID: ste.ParentID,
		RootID:   ste.RootID,
		Name:     db.NewNullString(ste.Name),
		IP:       ste.IP,
		Iss:      ste.Token.OIDCIssuer,
		Sub:      ste.Token.OIDCSubject,
	}
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		if ste.rtID == nil {
			if _, err := tx.Exec(`CALL CryptStoreRT_Insert(?,@ID)`, ste.rtEncrypted); err != nil {
				return errors.WithStack(err)
			}
			var rtID uint64
			if err := tx.Get(&rtID, `SELECT @ID`); err != nil {
				return errors.WithStack(err)
			}
			ste.rtID = &rtID
		}
		steStore.RefreshTokenID = *ste.rtID
		if err := steStore.Store(tx); err != nil {
			return err
		}
		if err := storeEncryptionKey(tx, ste.encryptionKeyEncrypted, steStore.RefreshTokenID, ste.ID); err != nil {
			return err
		}
		return eventService.LogEvent(tx, eventService.MTEvent{
			Event: event.FromNumber(event.MTEventCreated, comment),
			MTID:  ste.ID,
		}, ste.networkData)
	})
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
}

// Store stores the mytokenEntryStore in the database; if this is the first token for this user, the user is also added
// to the db
func (e *mytokenEntryStore) Store(tx *sqlx.Tx) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`CALL MTokens_Insert(?,?,?,?,?,?,?,?,?)`,
			e.Sub, e.Iss, e.ID, e.SeqNo, e.ParentID, e.RootID, e.RefreshTokenID, e.Name, e.IP)
		return errors.WithStack(err)
	})
}
