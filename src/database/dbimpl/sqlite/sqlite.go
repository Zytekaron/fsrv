package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"fsrv/src/database/dbimpl"
	"fsrv/src/database/entities"
)
import _ "github.com/mattn/go-sqlite3"
import _ "embed"

type SQLiteDB struct {
	db *sql.DB
	qm *QueryManager
}

const connParams = "?cache=shared&mode=rwc&_journal_mode=WAL"

/////////////////////////////////////////
//									   //
/*-------------------------------------*\
//         INTERFACE FUNCTIONS		   //
\*-------------------------------------*/
//									   //
/////////////////////////////////////////

/*               *\
	 Database
	 Functions
\*               */

//go:embed dbqueries/create.sql
var sqliteDatabaseCreationQuery string

func Create(databaseFile string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", databaseFile+connParams)
	if err != nil {
		return nil, err
	}

	sqliteDB := SQLiteDB{db, nil}

	_, err = db.Exec(sqliteDatabaseCreationQuery)
	if err != nil {
		return nil, err
	}

	sqliteDB.qm, err = NewQueryManager(db)
	if err != nil {
		return nil, err
	}

	return &sqliteDB, nil
}

func Open(databaseFile string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", databaseFile+connParams)
	if err != nil {
		return nil, err
	}

	qm, err := NewQueryManager(db)
	if err != nil {
		return nil, err
	}

	sqliteDB := SQLiteDB{db, qm}
	if err != nil {
		return nil, err
	}

	return &sqliteDB, nil
}

func Exists(databaseFile string) error {
	return dbimpl.Exists(databaseFile)
}

//go:embed dbqueries/check.sql
var sqliteCheckQuery string

func (sqlite *SQLiteDB) Check() error {
	rows, err := sqlite.db.Query(sqliteCheckQuery)
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
		"sqlite_master":    false,
	}

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

//go:embed dbqueries/destroy.sql
var sqliteDatabaseDestructionQuery string

func (sqlite *SQLiteDB) Destroy() error {
	_, err := sqlite.db.Query(sqliteDatabaseDestructionQuery)
	return err
}

/////////////////////////////////////////
//									   //
/*-------------------------------------*\
//       NON INTERFACE FUNCTIONS	   //
\*-------------------------------------*/
//									   //
/////////////////////////////////////////

// getResourceRoles
func (sqlite *SQLiteDB) getResourceRolePermIter(tx *sql.Tx, resourceID string) (func() error, *entities.RolePerm, error) {
	stmt := tx.Stmt(sqlite.qm.GetResourceRoles)
	rows, err := stmt.Query(resourceID)
	if err != nil {
		return nil, nil, err
	}

	var rolePerm entities.RolePerm
	roleIterNext := func() error {
		rows.Next()
		return rows.Scan(&rolePerm.Role, &rolePerm.Perm.Status, &rolePerm.Perm.TypeRWMD)
	}
	return roleIterNext, &rolePerm, nil
}

func (sqlite *SQLiteDB) createResourcePermissions(tx *sql.Tx, resource *entities.Resource) error {
	var err error
	var allowID int64 = -1
	var denyID int64 = -1

	for key, status := range resource.OperationNodes {
		if status && allowID == -1 {
			allowID, err = sqlite.constructPermNode(tx, &entities.Permission{
				ResourceID: resource.ID,
				TypeRWMD:   key.Type,
				Status:     status,
			})
			if err != nil {
				return err
			}
			err = sqlite.grantPermNode(tx, allowID, key.ID)
			if err != nil {
				return err
			}
		} else if denyID == -1 {
			denyID, err = sqlite.constructPermNode(tx, &entities.Permission{
				ResourceID: resource.ID,
				TypeRWMD:   key.Type,
				Status:     status,
			})
			if err != nil {
				return err
			}
			err = sqlite.grantPermNode(tx, denyID, key.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func rollbackOrPanic(tx *sql.Tx) {
	err := tx.Rollback()
	if err != nil {
		panic("BAD DATABASE STATE (TRANSACTION FAILED TO ROLL BACK): " + err.Error())
	}
}

func commitOrPanic(tx *sql.Tx) {
	err := tx.Commit()
	if err != nil {
		panic("BAD DATABASE STATE (TRANSACTION FAILED TO COMMIT): " + err.Error())
	}
}

// returns a string containing n query parameters
// useful for converting go arrays into multiple query parameters
// The format is the pattern that is repeated n times
// The last character in the paramFormat string is expected to be a "," so that it may be trimmed
func getNParams(paramFormat string, n int) string {
	if n < 1 {
		panic("getNParams must not recieve a query number < 1")
	}

	queryParams := ""
	for i := 0; i < n; i++ {
		queryParams += paramFormat
	}
	queryParams = queryParams[:len(queryParams)-1]

	return queryParams
}
