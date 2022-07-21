package database

import (
	"errors"
	"fsrv/src/config"
	"fsrv/src/database/dbimpl/sqlite"
)

func Create(cfg *config.Database) (DBInterface, error) {
	if cfg.Type == config.DatabaseSQLite {
		return sqlite.Create(cfg.Path)
	}
	return nil, errors.New("invalid database type")
}
