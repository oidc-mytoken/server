package dbModels

import (
	"github.com/jmoiron/sqlx"
)

// DBModel is an interface for database models
type DBModel interface {
	Store(tx *sqlx.Tx) error
}
