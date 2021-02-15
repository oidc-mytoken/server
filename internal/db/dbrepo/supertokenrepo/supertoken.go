package supertokenrepo

import (
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/utils/hashUtils"
	eventService "github.com/oidc-mytoken/server/shared/supertoken/event"
	event "github.com/oidc-mytoken/server/shared/supertoken/event/pkg"
	supertoken "github.com/oidc-mytoken/server/shared/supertoken/pkg"
	"github.com/oidc-mytoken/server/shared/supertoken/pkg/stid"
	"github.com/oidc-mytoken/server/shared/utils/cryptUtils"
)

// SuperTokenEntry holds the information of a SuperTokenEntry as stored in the
// database
type SuperTokenEntry struct {
	ID           stid.STID
	ParentID     stid.STID `db:"parent_id"`
	RootID       stid.STID `db:"root_id"`
	Token        *supertoken.SuperToken
	RefreshToken string `db:"refresh_token"`
	Name         string
	CreatedAt    time.Time `db:"created_at"`
	IP           string    `db:"ip_created"`
	networkData  model.ClientMetaData
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

func (ste *SuperTokenEntry) EncryptRT() (hash, crypt string, err error) {
	hash = hashUtils.SHA512Str([]byte(ste.RefreshToken))
	var key string
	key, err = ste.Token.ToJWT()
	if err != nil {
		return
	}
	crypt, err = cryptUtils.AES256Encrypt(ste.RefreshToken, key)
	return
}

// Store stores the SuperTokenEntry in the database
func (ste *SuperTokenEntry) Store(tx *sqlx.Tx, comment string) error {
	rtHash, rtCrypt, err := ste.EncryptRT()
	if err != nil {
		return err
	}
	steStore := superTokenEntryStore{
		ID:               ste.ID,
		ParentID:         ste.ParentID,
		RootID:           ste.RootID,
		RefreshToken:     rtCrypt,
		RefreshTokenHash: rtHash,
		Name:             db.NewNullString(ste.Name),
		IP:               ste.IP,
		Iss:              ste.Token.OIDCIssuer,
		Sub:              ste.Token.OIDCSubject,
	}
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		err = steStore.Store(tx)
		if err != nil {
			return err
		}
		return eventService.LogEvent(tx, event.FromNumber(event.STEventSTCreated, comment), ste.ID, ste.networkData)
	})
}

type superTokenEntryStore struct {
	ID               stid.STID
	ParentID         stid.STID `db:"parent_id"`
	RootID           stid.STID `db:"root_id"`
	RefreshToken     string    `db:"refresh_token"`
	RefreshTokenHash string    `db:"rt_hash"`
	Name             sql.NullString
	IP               string `db:"ip_created"`
	Iss              string
	Sub              string
}

// Store stores the superTokenEntryStore in the database; if this is the first token for this user, the user is also added to the db
func (e *superTokenEntryStore) Store(tx *sqlx.Tx) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		stmt, err := tx.PrepareNamed(`INSERT INTO SuperTokens (id, parent_id, root_id, refresh_token, rt_hash, name, ip_created, user_id) VALUES(:id, :parent_id, :root_id, :refresh_token, :rt_hash, :name, :ip_created, (SELECT id FROM Users WHERE iss=:iss AND sub=:sub))`)
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
