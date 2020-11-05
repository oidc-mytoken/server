package dbModels

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/zachmann/mytoken/internal/model"

	"github.com/go-sql-driver/mysql"

	"github.com/jmoiron/sqlx"

	"github.com/zachmann/mytoken/internal/supertoken/capabilities"
	"github.com/zachmann/mytoken/internal/supertoken/restrictions"

	eventService "github.com/zachmann/mytoken/internal/supertoken/event"
	event "github.com/zachmann/mytoken/internal/supertoken/event/pkg"

	"github.com/zachmann/mytoken/internal/db"

	uuid "github.com/satori/go.uuid"
	supertoken "github.com/zachmann/mytoken/internal/supertoken/pkg"
)

// SuperTokenEntry holds the information of a SuperTokenEntry as stored in the
// database
type SuperTokenEntry struct {
	ID           uuid.UUID
	ParentID     string `db:"parent_id"`
	RootID       string `db:"root_id"`
	Revoked      bool
	Token        *supertoken.SuperToken
	RefreshToken string `db:"refresh_token"`
	Name         string
	CreatedAt    time.Time `db:"created_at"`
	IP           string    `db:"ip_created"`
	networkData  model.NetworkData
}

func NewSuperTokenEntry(name, oidcSub, oidcIss string, r restrictions.Restrictions, c capabilities.Capabilities, networkData model.NetworkData) *SuperTokenEntry {
	st := supertoken.NewSuperToken(oidcSub, oidcIss, r, c)
	return &SuperTokenEntry{
		ID:          st.ID,
		Token:       st,
		Name:        name,
		IP:          networkData.IP,
		networkData: networkData,
	}
}

func (ste *SuperTokenEntry) Root() bool {
	if ste.RootID == "" {
		return true
	}
	return false
}

func (ste *SuperTokenEntry) Store(comment string) error {
	steStore := superTokenEntryStore{
		ID:           ste.ID,
		ParentID:     db.NewNullString(ste.ParentID),
		RootID:       db.NewNullString(ste.RootID),
		Revoked:      ste.Revoked,
		Token:        ste.Token,
		RefreshToken: db.NewNullString(ste.RefreshToken),
		Name:         db.NewNullString(ste.Name),
		IP:           ste.IP,
		Iss:          ste.Token.OIDCIssuer,
		Sub:          ste.Token.OIDCSubject,
	}
	err := steStore.Store()
	if err != nil {
		return err
	}
	return eventService.LogEvent(*event.FromNumber(event.STEventSTCreated, comment), ste.ID, ste.networkData)
}

type superTokenEntryStore struct {
	ID           uuid.UUID
	ParentID     sql.NullString `db:"parent_id"`
	RootID       sql.NullString `db:"root_id"`
	Revoked      bool
	Token        *supertoken.SuperToken
	RefreshToken sql.NullString `db:"refresh_token"`
	Name         sql.NullString
	IP           string `db:"ip_created"`
	Iss          string
	Sub          string
}

func (e *superTokenEntryStore) Store() error {
	stmt, err := db.DB().PrepareNamed(`INSERT INTO SuperTokens (id, parent_id, root_id, revoked, token, refresh_token, name, ip_created, user_id) VALUES(:id, :parent_id, :root_id, :revoked, :token, :refresh_token, :name, :ip_created, (SELECT id FROM Users WHERE iss=:iss AND sub=:sub))`)
	if err != nil {
		return err
	}
	return db.Transact(func(tx *sqlx.Tx) error {
		txStmt := tx.NamedStmt(stmt)
		_, err := txStmt.Exec(e)
		if err != nil {
			log.Printf("%s", err)
			var mysqlError *mysql.MySQLError
			if errors.As(err, &mysqlError) && mysqlError.Number == 1048 {
				_, err = tx.NamedExec(`INSERT INTO Users (sub, iss) VALUES(:sub, :iss)`, e)
				if err != nil {
					return err
				}
				_, err = txStmt.Exec(e)
				return err
			}
			return err
		}
		return nil
	})
}
