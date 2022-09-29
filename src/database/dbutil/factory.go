package dbutil

import (
	"errors"
	"fsrv/src/config"
	"fsrv/src/database"
	"fsrv/src/database/impl/sqlite"
	"io/fs"
	"os"
	"path/filepath"
)

func Create(cfg *config.Database) (database.DBInterface, error) {
	if cfg.Type == config.DatabaseSQLite {
		db, err := sqlite.Open(cfg.ConnectionString)
		if err != nil {
			err = createFile(cfg.ConnectionString)
			if err != nil {
				return nil, err
			}

			return sqlite.Create(cfg.ConnectionString)
		}
		return db, err
	}
	return nil, errors.New("invalid database type")
}

func createFile(path string) error {
	dir := filepath.Dir(path)
	if dir != "" {
		s, err := os.Stat(dir)
		if err != nil {
			err := os.MkdirAll(dir, 0760)
			if err != nil {
				return err
			}
		} else {
			// require at least 770 (rwxrwx---)
			if s.Mode()&0770 != 0770 {
				return fs.ErrPermission
			}
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}
