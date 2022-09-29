package dbutil

import (
	"errors"
	"fsrv/src/config"
	"fsrv/src/database"
	"fsrv/src/database/dbimpl/sqlite"
)

func Create(cfg *config.Database) (database.DBInterface, error) {
	if cfg.Type == config.DatabaseSQLite {
		db, err := sqlite.Open(cfg.Path)
		if err != nil {
			return sqlite.Create(cfg.Path)
		}
		return db, err
	}
	return nil, errors.New("invalid database type")
}
