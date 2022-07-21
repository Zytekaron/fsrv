package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"fsrv/src/database"
	"fsrv/src/database/dbimpl"
	"fsrv/src/database/entities"
	"fsrv/src/types"
	"fsrv/utils/serde"
	"log"
	"strconv"
	"time"
)
import _ "github.com/mattn/go-sqlite3"
import _ "embed"

type SQLiteDB struct {
	db *sql.DB
}

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

func (sqlite SQLiteDB) Create(databaseFile string) (database.DBInterface, error) {
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

func (sqlite SQLiteDB) Open(databaseFile string) (database.DBInterface, error) {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		return nil, err
	}
	return &SQLiteDB{db}, nil
}

func (sqlite SQLiteDB) Exists(databaseFile string) error {
	return dbimpl.Exists(databaseFile)
}

//go:embed dbqueries/check.sql
var sqliteCheckQuery string

func (sqlite SQLiteDB) Check() error {
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

//go:embed dbqueries/destroy.sql
var sqliteDatabaseDestructionQuery string

func (sqlite SQLiteDB) Destroy() error {
	_, err := sqlite.db.Query(sqliteDatabaseDestructionQuery)
	return err
}

func (sqlite SQLiteDB) CreateKey(key *entities.Key) error {
	if key.ID != "" {
		return errors.New("required feild keyid not specified")
	}

	//begin transaction
	_, err := sqlite.db.Query("BEGIN TRANSACTION;")
	if err != nil {
		return err
	}
	//create key record
	_, err = sqlite.db.Query("INSERT INTO Keys (keyid, note, expires, created) VALUES (?, ?, ?, ?)",
		key.ID, key.Comment, time.Time(key.ExpiresAt).UnixMilli(), time.Time(key.CreatedAt).UnixMilli())
	if err != nil {
		_, _ = sqlite.db.Query("ROLLBACK;")
		return err
	}

	//add RateLimit if exists
	if key.RequestRateLimit != nil {
		_, err = sqlite.db.Query("INSERT INTO Ratelimits (keyid, requests, reset) VALUES (?, ?, ?)",
			key.ID, key.RequestRateLimit.Limit, time.Duration(key.RequestRateLimit.Reset).Milliseconds())
		if err != nil {
			_, _ = sqlite.db.Query("ROLLBACK;")
			return err
		}
	}

	//add Roles
	var rows *sql.Rows
	var roleid int
	for _, role := range key.Roles {
		rows, err = sqlite.db.Query("SELECT roleid FROM Roles WHERE roleName = ?", role)
		if err != nil {
			_, _ = sqlite.db.Query("ROLLBACK;")
			return err
		}
		err = rows.Scan(roleid)
		if err != nil {
			_, _ = sqlite.db.Query("ROLLBACK;")
			return err
		}

		_, err = sqlite.db.Query("INSERT INTO KeyRoleIntersect (keyid, roleid) VALUES (?, ?)", key.ID, roleid)
		if err != nil {
			_, _ = sqlite.db.Query("ROLLBACK;")
			return err
		}
	}

	//add KeyRole //todo: ensure that precedence ordering is consistent
	_, err = sqlite.db.Query("INSERT INTO Roles (roleName, roleTypeRK, rolePrecedence) VALUES (?, 1, 10000)", key.ID)
	if err != nil {
		_, _ = sqlite.db.Query("ROLLBACK;")
		return err
	}

	//commit transaction results
	_, err = sqlite.db.Query("COMMIT;")
	if err != nil {
		_, _ = sqlite.db.Query("ROLLBACK;")
		return err
	}

	return nil
}

/*               *\
	 Object
	 Creation
\*               */

func (sqlite SQLiteDB) CreateResource(resource *entities.Resource) error {
	//begin transaction
	_, err := sqlite.db.Query("BEGIN TRANSACTION;")
	if err != nil {
		return err
	}

	//insert resource with flags
	_, err = sqlite.db.Query("INSERT INTO Resources (resourceid, flags) VALUES (?, ?)", resource.ID, resource.Flags)
	if err != nil {
		return err
	}

	//insert permissions
	err = sqlite.createResourcePermission(resource, resource.ReadNodes, types.OperationRead)
	if err != nil {
		return err
	}
	err = sqlite.createResourcePermission(resource, resource.WriteNodes, types.OperationWrite)
	if err != nil {
		return err
	}
	err = sqlite.createResourcePermission(resource, resource.ModifyNodes, types.OperationModify)
	if err != nil {
		return err
	}
	err = sqlite.createResourcePermission(resource, resource.DeleteNodes, types.OperationDelete)
	if err != nil {
		return err
	}

	//commit transaction results
	_, err = sqlite.db.Query("COMMIT;")
	if err != nil {
		_, _ = sqlite.db.Query("ROLLBACK;")
		return err
	}

	return nil
}

func (sqlite SQLiteDB) CreateRole(role *entities.Role) error {
	//note:roleTypeRK (0 = role, 1 = key)
	_, err := sqlite.db.Query("INSERT INTO Roles (roleName, rolePrecedence, roleTypeRK) VALUES (?, ?, 0)", role.ID, role.Precedence)
	return err
}

func (sqlite SQLiteDB) CreateRateLimit(keyid string, limit *entities.RateLimit) {
	//TODO implement me
	panic("implement me")
}

/*               *\
	 Data
	 Retrieval
\*               */

func (sqlite SQLiteDB) GetKeys() []*entities.Key {
	//TODO implement me
	panic("implement me")
}

func (sqlite SQLiteDB) GetKeyIDs() []string {
	//TODO implement me
	panic("implement me")
}

func (sqlite SQLiteDB) GetKeyData(keyid string) (*entities.Key, error) {
	//TODO implement me
	panic("implement me")
}

func (sqlite SQLiteDB) GetResources() []*entities.Resource {
	//TODO implement me
	panic("implement me")
}

func (sqlite SQLiteDB) GetResourceIDs() []string {
	//TODO implement me
	panic("implement me")
}

func (sqlite SQLiteDB) GetResourceData(resourceid string) (*entities.Resource, error) {
	var perm entities.Resource
	//get flags
	rows, err := sqlite.db.Query("SELECT flags FROM Resources where resourceid = ?", resourceid)
	if err != nil {
		return nil, err
	}

	err = rows.Scan(perm.ID)
	if err != nil {
		return nil, err
	}

	//get permission iterator
	iter, roleperm, err2 := sqlite.getResourceRolePermIter(resourceid)
	if err != nil {
		return nil, err2
	}

	//get permissions
	for iter() == nil {
		switch roleperm.TypeRW {
		case types.OperationRead:
			perm.ReadNodes[roleperm.Role.ID] = roleperm.Status
		case types.OperationWrite:
			perm.WriteNodes[roleperm.Role.ID] = roleperm.Status
		case types.OperationModify:
			perm.ModifyNodes[roleperm.Role.ID] = roleperm.Status
		case types.OperationDelete:
			perm.DeleteNodes[roleperm.Role.ID] = roleperm.Status
		default:
			//todo: make into error
			log.Println("[error] bad db state")
		}
	}

	return &perm, nil
}

func (sqlite SQLiteDB) GetRoles() []string {
	//TODO implement me
	panic("implement me")
}

/*               *\
	 Update
	 Functions
\*               */

func (sqlite SQLiteDB) GiveRole(keyid string, role ...string) error {
	//TODO implement me
	panic("implement me")
}

func (sqlite SQLiteDB) TakeRole(keyid string, role ...string) error {
	//TODO implement me
	panic("implement me")
}

func (sqlite SQLiteDB) GrantPermission(resource string, operationType types.OperationType, role ...string) []error {
	//TODO implement me
	panic("implement me")
}

func (sqlite SQLiteDB) RevokePermission(resource string, operationType types.OperationType, role ...string) []error {
	//TODO implement me
	panic("implement me")
}

func (sqlite SQLiteDB) SetRateLimit(keyid string, limit *entities.RateLimit) {
	//TODO implement me
	panic("implement me")
}

/*               *\
	 Object
	 Deletion
\*               */

func (sqlite SQLiteDB) DeleteRole(name string) error {
	//TODO implement me
	panic("implement me")
}

func (sqlite SQLiteDB) DeleteKey(id string) error {
	//TODO implement me
	panic("implement me")
}

func (sqlite SQLiteDB) DeleteResource(id string) error {
	//TODO implement me
	panic("implement me")
}

/////////////////////////////////////////
//									   //
/*-------------------------------------*\
//       NON INTERFACE FUNCTIONS	   //
\*-------------------------------------*/
//									   //
/////////////////////////////////////////

//go:embed getRateLimit.sql
var sqliteGetRateLimit string

// todo: REMOVE UNUSED
// getRateLimit returns a RateLimit object for a given key
func (sqlite *SQLiteDB) getRateLimit(keyid string) (*entities.RateLimit, error) {
	rows, err := sqlite.db.Query(sqliteGetRateLimit)
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

//go:embed getResourceRoles.sql
var sqliteGetResourceRoles string

// getResourceRoles
func (sqlite *SQLiteDB) getResourceRolePermIter(resourceID string) (func() error, *entities.RolePerm, error) {
	rows, err := sqlite.db.Query(sqliteGetResourceRoles, resourceID)
	if err != nil {
		return nil, nil, err
	}

	var rolePerm entities.RolePerm
	roleIterNext := func() error {
		return rows.Scan(rolePerm.Role, rolePerm.Status, rolePerm.TypeRW)
	}
	return roleIterNext, &rolePerm, nil
}

func (sqlite *SQLiteDB) createResourcePermission(resource *entities.Resource, permMap map[string]bool, operationType types.OperationType) error {
	size := len(permMap)

	iters := 0
	query := ""
	rolesAndPerms := make([]string, 0, size*3)
	for _, status := range permMap {
		iters++
		query += "(?, ?, ?)"
		if iters < size {
			query += ", "
		}

		//todo: check if strconv is the best idea for this (implement method for operationType?)
		rolesAndPerms = append(rolesAndPerms, strconv.Itoa(int(operationType)))
		rolesAndPerms = append(rolesAndPerms, resource.ID)
		if status {
			rolesAndPerms = append(rolesAndPerms, "1")
		} else {
			rolesAndPerms = append(rolesAndPerms, "0")
		}
	}

	_, err := sqlite.db.Query("INSERT INTO Permissions (resourceid, permTypeRWMD, permTypeDenyAllow) VALUES "+query, rolesAndPerms)
	if err != nil {
		return err
	}

	return nil
}
