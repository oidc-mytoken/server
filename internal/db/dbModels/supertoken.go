package dbModels

import (
	"database/sql"
	"time"

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
}

func NewSuperTokenEntry(name, oidcSub, oidcIss string, r restrictions.Restrictions, c capabilities.Capabilities) *SuperTokenEntry {
	//TODO
	ip := "192.168.0.31"

	st := supertoken.NewSuperToken(oidcSub, oidcIss, r, c)
	return &SuperTokenEntry{
		ID:    st.ID,
		Token: st,
		Name:  name,
		IP:    ip,
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
	return eventService.LogEvent(*event.FromNumber(event.STEventSTCreated, comment), ste.ID)
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
	_, err := db.DB().NamedExec(`INSERT INTO SuperTokens (id, parent_id, root_id, revoked, token, refresh_token, name, ip_created, user_id) VALUES(:id, :parent_id, :root_id, :revoked, :token, :refresh_token, :name, :ip_created, (SELECT id FROM Users WHERE iss=:iss AND sub=:sub))`, e)
	return err
}
