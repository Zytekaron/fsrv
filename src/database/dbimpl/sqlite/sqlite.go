package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"fsrv/src/database/entities"
	"fsrv/src/types"
	"fsrv/utils/serde"
	"log"
	"os"
	"time"
)
import _ "github.com/mattn/go-sqlite3"
import _ "embed"

type SQLiteDB struct {
	db *sql.DB
}

// New creates an SQLiteDB object by opening the database
func New(databaseFile string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		return nil, err
	}
	return &SQLiteDB{db}, nil
}

// Exists checks if a database file exists, is readable, and is writable
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

// Create makes a new database along with the necessary tables and indexes
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

// Check performs a database integrity check
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

// Destroy runs a query that destroys all database objects made in the Create function
func (db *SQLiteDB) Destroy(databaseFile string) error {
	_, err := db.db.Query(sqliteDatabaseDestructionQuery)
	return err
}

//go:embed getRateLimit.sql
var sqliteGetRateLimit string

// getRateLimit returns a RateLimit object for a given key
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
	}

	return &entities.RateLimit{
		Limit: requests,
		Reset: serde.Duration(reset * int64(time.Millisecond)),
	}, nil
}

//go:embed getResourceFlags.sql
var sqliteGetResourcePublicStatus string

// isResourcePublic returns true if a resource is publicly available for reads
func (db *SQLiteDB) isResourcePublic(resourceID string) (bool, error) {
	rows, err := db.db.Query(sqliteGetResourcePublicStatus, resourceID)
	if err != nil {
		return false, err
	}

	var public bool
	err = rows.Scan(public)
	if err != nil {
		return false, err
	}

	return public, nil
}

//go:embed getResourceRoles.sql
var sqliteGetResourceRoles string

// getResourceRoles
func (db *SQLiteDB) getResourceRolePermIter(resourceID string) (func() error, *entities.RolePerm, error) {
	rows, err := db.db.Query(sqliteGetResourceRoles, resourceID)
	if err != nil {
		return nil, nil, err
	}

	var rolePerm entities.RolePerm
	roleIterNext := func() error {
		return rows.Scan(rolePerm.Role, rolePerm.Status, rolePerm.TypeRW)
	}
	return roleIterNext, &rolePerm, nil
}

func (db *SQLiteDB) getResourcePermission(resourceID string) (*entities.Resource, error) {
	var perm entities.Resource
	rows, err := db.db.Query("SELECT flags FROM Resources where resourceid = ?", resourceID)
	if err != nil {
		return nil, err
	}

	err = rows.Scan(perm.ID)
	if err != nil {
		return nil, err
	}

	iter, roleperm, err2 := db.getResourceRolePermIter(resourceID)
	if err != nil {
		return nil, err2
	}

	for iter() == nil {
		switch roleperm.TypeRW {
		case types.OperationRead:
			perm.ReadNodes[roleperm.Role] = roleperm.Status
		case types.OperationWrite:
			perm.WriteNodes[roleperm.Role] = roleperm.Status
		case types.OperationModify:
			perm.ModifyNodes[roleperm.Role] = roleperm.Status
		case types.OperationDelete:
			perm.DeleteNodes[roleperm.Role] = roleperm.Status
		default:
			//todo: make into error
			log.Println("[error] bad db state")
		}
	}

	return &perm, nil
}

func (db *SQLiteDB) setRateLimit(keyid string, rateLimit entities.RateLimit) {

}
