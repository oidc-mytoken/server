package dbModels

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
	"github.com/zachmann/mytoken/internal/db"
	"github.com/zachmann/mytoken/internal/supertoken/capabilities"
)

type AccessToken struct {
	Token   string
	IP      string `db:"ip_created"`
	Comment string
	STID    uuid.UUID `db:"ST_id"`

	Scopes       []string
	Capabilities capabilities.Capabilities
	Audiences    []string
}

type accessToken struct {
	Token   string
	IP      string
	Comment sql.NullString
	STID    uuid.UUID

	ScopeAttr      []Attribute
	CapabilityAttr []Attribute
	AudienceAttr   []Attribute
}

func (t *AccessToken) Store() error {
	return db.Transact(func(tx *sqlx.Tx) error {
		res, err := tx.NamedExec(`INSERT INTO AccessTokens (token, ip_created, comment, ST_id) VALUES (:token, :ip_created, :comment, :ST_id)`, t)
		if err != nil {
			return err
		}
		atID, err := res.LastInsertId()
		if err != nil {
			return err
		}

		if _, err := tx.NamedExec(`INSERT INTO AT_Attributes (AT_id, attribute_id, attribute) VALUES (:token, :ip_created, :comment, :ST_id)`, t); err != nil {
			return err
		}

		return nil
	})
}
