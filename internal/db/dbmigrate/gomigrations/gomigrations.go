package gomigrations

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/versionrepo"
)

type updateFunc = func(tx *sqlx.Tx) error

var migrations = make(map[string]updateFunc)

func RunGoMigration(tx *sqlx.Tx, version string) error {
	fnc, ok := migrations[version]
	if !ok {
		fmt.Printf("No Go migrations for %s\n", version)
		return nil
	}
	fmt.Printf("Running Go migrations for %s\n", version)
	return db.RunWithinTransaction(
		log.StandardLogger(), tx, func(tx *sqlx.Tx) error {
			if err := fnc(tx); err != nil {
				return errors.WithStack(err)
			}
			return versionrepo.SetVersionGo(log.StandardLogger(), tx, version)
		},
	)
}
