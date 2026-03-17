package mytokenrepo

import (
	"database/sql"
	"encoding/base64"

	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/unixtime"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	helper "github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	"github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/utils/cryptutils"
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
	Tags                   []api.CreateMytokenTag
}

// InitRefreshToken links a refresh token to this MytokenEntry
func (mte *MytokenEntry) InitRefreshToken(rt string) (err error) {
	mte.refreshToken = rt
	mte.encryptionKey, err = cryptutils.RandomBytes(32)
	if err != nil {
		return
	}
	tmp, err := cryptutils.AESEncrypt(mte.refreshToken, mte.encryptionKey)
	if err != nil {
		return err
	}
	mte.rtEncrypted = tmp
	jwt, err := mte.Token.ToJWT()
	if err != nil {
		return err
	}
	tmp, err = cryptutils.AES256Encrypt(base64.StdEncoding.EncodeToString(mte.encryptionKey), jwt)
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
	tmp, err := cryptutils.AES256Encrypt(base64.StdEncoding.EncodeToString(key), jwt)
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
	meta, err := mte.Token.DBMetadata()
	if err != nil {
		return err
	}
	steStore.MytokenDBMetadata = meta

	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			if mte.rtID == nil {
				if _, err = tx.Exec(`CALL CryptStoreRT_Insert(?,@ID)`, mte.rtEncrypted); err != nil {
					return errors.WithStack(err)
				}
				var rtID uint64
				if err = tx.Get(&rtID, `SELECT @ID`); err != nil {
					return errors.WithStack(err)
				}
				mte.rtID = &rtID
			}
			steStore.RefreshTokenID = *mte.rtID
			if err = steStore.Store(rlog, tx); err != nil {
				return err
			}
			if err = storeEncryptionKey(tx, mte.encryptionKeyEncrypted, steStore.RefreshTokenID, mte.ID); err != nil {
				return err
			}
			if err = eventService.LogEvent(
				rlog, tx, pkg.MTEvent{
					Event:          api.EventMTCreated,
					Comment:        comment,
					MTID:           mte.ID,
					ClientMetaData: mte.networkData,
				},
			); err != nil {
				return err
			}
			// Link tags to the newly created mytoken
			for _, tag := range mte.Tags {
				if _, err = tx.Exec(
					`CALL MTokens_LinkTag(?,?,?)`, mte.ID, tag.Tag, tag.IncludeChildren,
				); err != nil {
					return errors.WithStack(err)
				}
			}
			return nil
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
	helper.MytokenDBMetadata
}

// Store stores the mytokenEntryStore in the database; if this is the first token for this user, the user is also added
// to the db
func (e *mytokenEntryStore) Store(rlog log.Ext1FieldLogger, tx *sqlx.Tx) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(
				`CALL MTokens_Insert(?,?,?,?,?,?,?,?,?,?,?,?)`,
				e.Sub, e.Iss, e.ID, e.SeqNo, e.ParentID, e.RefreshTokenID, e.Name, e.IP, e.ExpiresAt, e.Capabilities,
				e.Rotation, e.Restrictions,
			)
			return errors.WithStack(err)
		},
	)
}

// AddTag adds a tag to a mytoken
func AddTag(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MOMID,
	tag api.Tag, includeChildren bool,
) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			ids := []mtid.MOMID{mtID}
			if includeChildren {
				var tokens []*MytokenEntry
				if err := db.RunWithinTransaction(
					rlog, tx, func(tx *sqlx.Tx) error {
						return errors.WithStack(
							tx.Select(
								&tokens,
								`CALL MTokens_GetSubtokens(?)`, mtID,
							),
						)
					},
				); err != nil {
					return err
				}
				for _, token := range tokens {
					ids = append(ids, token.ID.MomID())
				}
			}
			for _, id := range ids {
				_, err := tx.Exec(`CALL MTokens_LinkTag(?,?,?)`, id, tag, includeChildren)
				if err != nil {
					return err
				}
			}
			return nil
		},
	)

}

// RemoveTag removes a tag from a mytoken
func RemoveTag(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MOMID, tag api.Tag,
) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			var includeChildren db.BitBool
			if err := tx.Get(&includeChildren, `CALL MTTag_GetIncludeChildren(?,?)`, mtID, tag); err != nil {
				return errors.WithStack(err)
			}
			ids := []mtid.MOMID{mtID}
			if includeChildren {
				var tokens []*MytokenEntry
				if err := errors.WithStack(
					tx.Select(&tokens, `CALL MTokens_GetSubtokens(?)`, mtID),
				); err != nil {
					return err
				}
				for _, token := range tokens {
					ids = append(ids, token.ID.MomID())
				}
			}
			for _, id := range ids {
				_, err := tx.Exec(`CALL MTokens_UnlinkTag(?,?)`, id, tag)
				if err != nil {
					return err
				}
			}
			return nil
		},
	)
}

// ClearTags removes all tags from a mytoken
func ClearTags(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MOMID,
) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			var children []*MytokenEntry
			if err := errors.WithStack(
				tx.Select(&children, `CALL MTokens_GetChildren(?)`, mtID),
			); err != nil {
				return err
			}

			var tags []struct {
				TagID              uint64     `db:"tag_id"`
				Tag                string     `db:"tag"`
				Color              string     `db:"tag_color"`
				TagIncludeChildren db.BitBool `db:"tag_include_children"`
			}
			if err := errors.WithStack(
				tx.Select(&tags, `CALL MTokens_GetTags(?)`, mtID),
			); err != nil {
				return err
			}

			for _, tag := range tags {
				var includeChildren bool
				if err := tx.Get(
					&includeChildren, `CALL MTTag_GetIncludeChildren(
?,?)`, mtID, tag,
				); err != nil {
					return errors.WithStack(err)
				}
				if includeChildren {
					for _, child := range children {
						if err := RemoveTag(rlog, tx, child.ID.MomID(), api.Tag(tag.Tag)); err != nil {
							return err
						}
					}
				}
			}
			_, err := tx.Exec(`CALL MTokens_ClearTags(?)`, mtID)
			return errors.WithStack(err)
		},
	)
}
