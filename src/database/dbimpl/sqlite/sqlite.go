package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"fsrv/src/database/entities"
	"fsrv/utils/serde"
	"os"
	"time"
)
import _ "github.com/mattn/go-sqlite3"
import _ "embed"

type SQLiteDB struct {
	db *sql.DB
}

func New(databaseFile string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		return nil, err
	}
	return &SQLiteDB{db}, nil
}

func Exists(databaseFile string) (bool, error) {
	_, err := os.Stat(databaseFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		} else {
			return false, err
		}
	}

	f, err := os.OpenFile(databaseFile, os.O_RDWR, 0666)
	if err != nil {
		return false, err
	}
	defer f.Close()
	return true, nil
}

//go:embed create.sql
var sqliteDatabaseCreationQuery string

func Create(databaseFile string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		return nil, err
	}

	_, err = db.Query(sqliteDatabaseCreationQuery)

	if err != nil {
		return nil, err
	}
	return &SQLiteDB{db}, nil
}

//go:embed check.sql
var sqliteCheckQuery string

func (db *SQLiteDB) Check() error {
	rows, err := db.db.Query(sqliteCheckQuery)
	if err != nil {
		return err
	}
	tableMap := map[string]bool{
		"KeyRoleIntersect": false,
		"Keys":             false,
		"Permissions":      false,
		"Ratelimits":       false,
		"Resources":        false,
		"KeyPermIntersect": false,
		"Roles":            false,
		"sqlite_master":    false}

	var name string
	for rows.Next() {
		err = rows.Scan(name)
		if err != nil {
			return err
		}
		if _, ok := tableMap[name]; ok {
			tableMap[name] = true
		} else {
			return errors.New(fmt.Sprintf("Extraneous table \"%s\" should not exist in database", name))
		}
	}

	for key, val := range tableMap {
		if !val {
			return errors.New(fmt.Sprintf("The table \"%s\" does not exist in database", key))
		}
	}

	return nil
}

//go:embed destroy.sql
var sqliteDatabaseDestructionQuery string

func (db *SQLiteDB) Destroy(databaseFile string) error {
	_, err := db.db.Query(sqliteDatabaseDestructionQuery)
	return err
}

//go:embed getRateLimitByKeyID.sql
var sqliteGetRateLimit string

func (db *SQLiteDB) getRateLimit(keyid string) (*entities.RateLimit, error) {
	rows, err := db.db.Query(sqliteGetRateLimit)
	var requests int
	var reset int64
	if err != nil {
		return nil, err
	}
	err = rows.Scan(requests, reset)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		return nil, errors.New("multiple rate limits exist for one key")
	} else {
		return &entities.RateLimit{
			ID:    keyid,
			Limit: requests,
			Reset: serde.Duration(reset * int64(time.Millisecond)),
		}, nil
	}
}
