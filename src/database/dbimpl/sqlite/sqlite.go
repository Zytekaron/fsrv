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

func Create(databaseFile string) (database.DBInterface, error) {
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

func Open(databaseFile string) (database.DBInterface, error) {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		return nil, err
	}
	return &SQLiteDB{db}, nil
}

func Exists(databaseFile string) error {
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
		_, err = sqlite.db.Query("INSERT INTO Ratelimits (ratelimitid, requests, reset) VALUES (?, ?, ?)",
			key.RequestRateLimit.ID, key.RequestRateLimit.Limit, time.Duration(key.RequestRateLimit.Reset).Milliseconds())
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

func (sqlite SQLiteDB) CreateRateLimit(limit *entities.RateLimit) error {
	_, err := sqlite.db.Query("INSERT INTO Ratelimits (ratelimitid, requests, reset) VALUES (?, ?, ?)", limit.ID, limit.Limit, limit.Reset)
	return err
}

/*               *\
	 Data
	 Retrieval
\*               */

func (sqlite SQLiteDB) GetKeys(pageSize int, offset int) ([]*entities.Key, error) {
	var keys []*entities.Key
	keyIDs, err := sqlite.GetKeyIDs(pageSize, offset)
	if err != nil {
		return keys, err
	}
	for i, keyID := range keyIDs {
		keys[i], err = sqlite.GetKeyData(keyID)
		if err != nil {
			return keys, nil
		}
	}
	return keys, nil
}

func (sqlite SQLiteDB) GetKeyIDs(pageSize int, offset int) ([]string, error) {
	keyIDs := make([]string, 0, pageSize)
	rows, err := sqlite.db.Query("SELECT keyid FROM Keys LIMIT ? OFFSET ?", pageSize, offset)

	for i := range keyIDs {
		err = rows.Scan(keyIDs[i])
		if err != nil {
			return keyIDs, err
		}
	}

	return keyIDs, err
}

func (sqlite SQLiteDB) GetKeyData(keyid string) (*entities.Key, error) {
	var key *entities.Key = nil
	var createMS, expireMS int64
	keyRows, err := sqlite.db.Query("SELECT note, ratelimitid, created, expires FROM Keys WHERE keyid = ?", keyid)
	if err != nil {
		return key, err
	}

	//todo: consider combining with main key query using join
	var rateLimitID string
	var rateLimit entities.RateLimit
	rateRows, err := sqlite.db.Query("SELECT ratelimitid, requests, reset FROM Ratelimits WHERE ratelimitid = ?", rateLimitID)
	if err != nil {
		return key, err
	}

	//todo: get roles by precedence using KeyRoleIntersect
	var roles []string
	var role string
	roleRows, err := sqlite.db.Query("SELECT roleName FROM Roles JOIN KeyRoleIntersect KRI on Roles.roleid = KRI.roleid WHERE keyid = ? ORDER BY rolePrecedence", keyid)
	err = rateRows.Scan(rateLimit.ID, rateLimit.Limit, rateLimit.Reset)
	if err != nil {
		return key, err
	}
	err = keyRows.Scan(key.Comment, rateLimitID, createMS, expireMS)
	if err != nil {
		return key, err
	}
	for roleRows.Next() {
		err = roleRows.Scan(role)
		if err != nil {
			return key, err
		}
		roles = append(roles, role)
	}

	key.ID = keyid
	key.CreatedAt = serde.Time(time.UnixMilli(createMS))
	key.ExpiresAt = serde.Time(time.UnixMilli(expireMS))
	key.RequestRateLimit = &rateLimit
	key.Roles = roles //todo: add roles to key struct

	return key, nil
}

func (sqlite SQLiteDB) GetResources(pageSize int, offset int) ([]*entities.Resource, error) {
	resourceIDs, err := sqlite.GetResourceIDs(pageSize, offset)
	if err != nil {
		return nil, nil
	}
	resources := make([]*entities.Resource, 0, len(resourceIDs))

	for i, id := range resourceIDs {
		resources[i], err = sqlite.GetResourceData(id)
		if err != nil {
			return resources, err
		}
	}

	return resources, nil
}

func (sqlite SQLiteDB) GetResourceIDs(pageSize int, offset int) ([]string, error) {
	resourceIDs := make([]string, 0, pageSize)
	arrPos := 0
	rows, err := sqlite.db.Query("SELECT resourceid FROM Resources LIMIT ? OFFSET ?", pageSize, offset)
	if err != nil {
		return resourceIDs, err
	}

	for rows.Next() {
		err = rows.Scan(resourceIDs[arrPos])
		if err != nil {
			return resourceIDs, err
		}
	}

	return resourceIDs, nil
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

func (sqlite SQLiteDB) GetRoles() ([]string, error) {
	var role string
	var roles []string
	rows, err := sqlite.db.Query("SELECT roleName FROM Roles where roleTypeRK=0")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		err = rows.Scan(role)
		if err != nil {
			return roles, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

/*               *\
	 Update
	 Functions
\*               */

func (sqlite SQLiteDB) GiveRole(keyid string, roles ...string) error {
	query := ""
	params := make([]string, len(roles)*2)
	for i, role := range roles {
		query += "(?, ?),"

		params[i*2] = keyid
		params[i*2+1] = role
	}

	query = query[:len(query)-1]

	_, err := sqlite.db.Query("INSERT INTO KeyRoleIntersect VALUES "+query, params)
	if err != nil {
		return err
	}
	return nil
}

func (sqlite SQLiteDB) TakeRole(keyid string, roles ...string) error {
	query := ""
	params := make([]string, len(roles))
	for i, role := range roles {
		query += "?,"

		params[i] = role
	}
	query = query[:len(query)-1]

	_, err := sqlite.db.Query("DELETE FROM KeyRoleIntersect WHERE keyid = ? AND roleid IN ("+query+")", keyid, params)
	if err != nil {
		return err
	}
	return nil
}

func (sqlite SQLiteDB) GrantPermission(resourceID string, operationType types.OperationType, denyAllow bool, roles ...string) []error {
	var errs []error
	var permissionID int
	//begin transaction
	tx, err := sqlite.db.Begin()
	if err != nil {
		errs = append(errs, errors.New("cannot begin transaction"))
		return errs
	}

	//get permissionid of existing/new permission node
	row := tx.QueryRow("SELECT permissionid FROM Permissions WHERE resourceid = ? AND permTypeRWMD = ? AND permTypeDenyAllow = ?", resourceID, operationType, denyAllow)
	err = row.Scan(permissionID)
	if err == sql.ErrNoRows {
		_, err = tx.Query("INSERT INTO Permissions (resourceid, permTypeRWMD, permTypeDenyAllow) VALUES (?, ?, ?)", resourceID, operationType, denyAllow)
		if err != nil {
			errs = append(errs, err)
			rollbackOrPanic(tx)
			return errs
		}
		row = tx.QueryRow("SELECT last_insert_rowid()")
		if err != nil {
			errs = append(errs, err)
			rollbackOrPanic(tx)
			return errs
		}
	}

	//add roles to permission node
	query := getNParams("(?,?),", len(roles))
	params := make([]string, len(roles)*2)
	for i, role := range roles {
		params[i*2] = role
		params[i*2+1] = strconv.Itoa(permissionID)
	}
	_, err = tx.Query("INSERT INTO RolePermIntersect (roleid, permissionid) VALUES"+query, params)
	if err != nil {
		errs = append(errs, err)
		rollbackOrPanic(tx)
		return errs
	}

	return nil
}

func (sqlite SQLiteDB) RevokePermission(resourceID string, operationType types.OperationType, denyAllow bool, roles ...string) error {
	/*Begin transaction that:
	-Retrieves a permissionID for the given resourceid, operationType, and denyAllow status
	-Removes the specified role(s) from the RolePermIntersect
	-Checks if any roles are still associated with that permission node
	-Removes the permission node by id if it has no associated roles
	*/
	tx, e := sqlite.db.Begin()
	if e != nil {
		return e
	}

	var permissionID string

	//get permission that connects to roles that need their connection to this permission revoked
	row := tx.QueryRow("SELECT permissionid FROM Permissions WHERE resourceid = ? AND permTypeRWMD = ? AND permTypeDenyAllow = ?", resourceID, operationType, denyAllow)
	err := row.Scan(permissionID)
	if err != nil {
		commitOrPanic(tx)
		return err
	}

	//delete RolePermIntersect entries
	params := getNParams("?,", len(roles))
	_, err = tx.Query("DELETE FROM RolePermIntersect WHERE permissionid = ? AND roleid IN ("+params+")", permissionID, roles)
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}

	//count remaining references
	row = tx.QueryRow("SELECT COUNT(permissionid) FROM RolePermIntersect WHERE permissionid = ?", permissionID)
	var references int
	err = row.Scan(references)
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}
	if references < 1 { //if no permissions reference the given permission
		_, err = tx.Query("DELETE FROM Permissions WHERE permissionid = ?", permissionID) //delete orphaned permission node
		if err != nil {
			rollbackOrPanic(tx)
			return err
		}
	}

	commitOrPanic(tx)

	return nil
}

func (sqlite SQLiteDB) SetRateLimit(key *entities.Key, limit *entities.RateLimit) error {
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

//todo: REMOVE UNUSED
//getRateLimit returns a RateLimit object for a given key
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

//returns a string containing n query parameters
//useful for converting go arrays into multiple query parameters
//The format is the pattern that is repeated n times
//The last character in the paramFormat string is expected to be a "," so that it may be trimmed
func getNParams(paramFormat string, n int) string {
	queryParams := ""
	for i := 0; i < n; i++ {
		queryParams += paramFormat
	}
	queryParams = queryParams[:len(queryParams)-1]

	return queryParams
}
