package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"fsrv/src/database/dberr"
	"fsrv/src/database/dbimpl"
	"fsrv/src/database/entities"
	"fsrv/src/types"
	"fsrv/utils/serde"
	"log"
	"time"
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

func (sqlite *SQLiteDB) CreateKey(key *entities.Key) error {
	if key.ID == "" {
		return errors.New("required feild keyid not specified")
	}

	//begin transaction
	tx, err := sqlite.db.Begin()
	if err != nil {
		return err
	}

	//create key record
	stmt := tx.Stmt(sqlite.qm.InsKeyData)
	_, err = stmt.Exec(key.ID, key.Comment, key.RateLimitID, time.Time(key.ExpiresAt).UnixMilli(), time.Time(key.CreatedAt).UnixMilli())
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}

	//add Roles
	var roleid string
	stmtGetRoleID := tx.Stmt(sqlite.qm.GetRoleIDIfExists)
	stmtInsKRIData := tx.Stmt(sqlite.qm.InsKeyRoleIntersectData)
	for _, role := range key.Roles {
		//check if key role exists
		row := stmtGetRoleID.QueryRow(role)
		err = row.Scan(&roleid) //produces sql.ErrNoRows if role does not exist
		if err != nil {
			rollbackOrPanic(tx)
			return err
		}

		//insert Role into KeyRoleIntersect //todo:make function
		_, err = stmtInsKRIData.Exec(key.ID, roleid)
		if err != nil {
			rollbackOrPanic(tx)
			return err
		}
	}

	//insert KeyRole into roles //todo: ensure that precedence ordering is consistent
	stmt = tx.Stmt(sqlite.qm.InsRoleData)
	_, err = stmt.Exec(key.ID, 1, 10000)
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}
	//Insert KeyRole into KeyRoleIntersect
	stmt = tx.Stmt(sqlite.qm.InsKeyRoleIntersectData)
	_, err = stmt.Exec(key.ID, key.ID)
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}

	//commit transaction results
	commitOrPanic(tx)

	return nil
}

/*               *\
	 Object
	 Creation
\*               */

func (sqlite *SQLiteDB) CreateResource(resource *entities.Resource) error {
	//begin transaction
	tx, err := sqlite.db.Begin()
	if err != nil {
		return err
	}

	//insert resource with flags
	stmt := tx.Stmt(sqlite.qm.InsResourceData)
	_, err = stmt.Exec(resource.ID, resource.Flags)
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}

	//insert permissions
	err = sqlite.createResourcePermission(tx, resource, resource.ReadNodes, types.OperationRead)
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}
	err = sqlite.createResourcePermission(tx, resource, resource.WriteNodes, types.OperationWrite)
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}
	err = sqlite.createResourcePermission(tx, resource, resource.ModifyNodes, types.OperationModify)
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}
	err = sqlite.createResourcePermission(tx, resource, resource.DeleteNodes, types.OperationDelete)
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}

	//commit transaction results
	commitOrPanic(tx)

	return nil
}

func (sqlite *SQLiteDB) CreateRole(role *entities.Role) error {
	tx, err := sqlite.db.Begin()
	if err != nil {
		return err
	}
	stmt := tx.Stmt(sqlite.qm.InsRoleData)
	res, err := stmt.Exec(role.ID, 0, role.Precedence)
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}
	rowsInserted, err := res.RowsAffected()
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}
	if rowsInserted == 0 {
		rollbackOrPanic(tx)
		return sql.ErrNoRows
	}
	commitOrPanic(tx)
	return err
}

func (sqlite *SQLiteDB) CreateRateLimit(limit *entities.RateLimit) error {
	_, err := sqlite.qm.InsRateLimitData.Exec(limit.ID, limit.Limit, limit.Reset)
	return err
}

/*               *\
	 Data
	 Retrieval
\*               */

func (sqlite *SQLiteDB) GetKeys(pageSize int, offset int) ([]*entities.Key, error) {
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

func (sqlite *SQLiteDB) GetKeyIDs(pageSize int, offset int) ([]string, error) {
	//todo: determine if this is okay to do outside of a transaction
	var keyIDs []string
	var id string
	rows, err := sqlite.qm.GetKeyIDs.Query(pageSize, offset)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return keyIDs, err
		}
		keyIDs = append(keyIDs, id)
	}

	return keyIDs, nil
}

func (sqlite *SQLiteDB) GetKeyData(keyid string) (*entities.Key, error) {
	var key entities.Key
	var rtlimID sql.NullString
	var createMS, expireMS int64

	//begin transaction
	tx, err := sqlite.db.Begin()
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	//get base key data
	stmtGetBaseData := tx.Stmt(sqlite.qm.GetKeyData)
	row := stmtGetBaseData.QueryRow(keyid)
	err = row.Scan(&key.Comment, &rtlimID, &createMS, &expireMS)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	//get Roles
	stmtGetRoles := tx.Stmt(sqlite.qm.GetRolesByKeyIDOrdered)
	rows, err := stmtGetRoles.Query(keyid)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	var role string
	for rows.Next() {
		err = rows.Scan(&role)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		key.Roles = append(key.Roles, role)
	}

	//finish building key
	key.ID = keyid
	key.CreatedAt = serde.Time(time.UnixMilli(createMS))
	key.ExpiresAt = serde.Time(time.UnixMilli(expireMS))

	commitOrPanic(tx)

	return &key, nil
}

func (sqlite *SQLiteDB) GetResources(pageSize int, offset int) ([]*entities.Resource, error) {
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

func (sqlite *SQLiteDB) GetResourceIDs(pageSize int, offset int) ([]string, error) {
	var resourceIDs []string
	var id string
	rows, err := sqlite.db.Query("SELECT resourceid FROM Resources LIMIT ? OFFSET ?", pageSize, offset)
	if err != nil {
		return resourceIDs, err
	}

	for rows.Next() {
		err = rows.Scan(&id)
		resourceIDs = append(resourceIDs, id)
		if err != nil {
			return resourceIDs, err
		}
	}

	return resourceIDs, nil
}

func (sqlite *SQLiteDB) GetResourceData(resourceid string) (*entities.Resource, error) {
	//begin transaction
	tx, err := sqlite.db.Begin()
	if err != nil {
		return nil, err
	}

	var res entities.Resource

	//get flags
	stmt := tx.Stmt(sqlite.qm.GetResourceFlagsByID)
	row := stmt.QueryRow(resourceid)

	err = row.Scan(&res.Flags)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	//get permission iterator
	iter, roleperm, err := sqlite.getResourceRolePermIter(tx, resourceid)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	//get permissions
	for iter() == nil {
		switch roleperm.Perm.TypeRWMD {
		case types.OperationRead:
			res.ReadNodes[roleperm.Role.ID] = roleperm.Perm.Status
		case types.OperationWrite:
			res.WriteNodes[roleperm.Role.ID] = roleperm.Perm.Status
		case types.OperationModify:
			res.ModifyNodes[roleperm.Role.ID] = roleperm.Perm.Status
		case types.OperationDelete:
			res.DeleteNodes[roleperm.Role.ID] = roleperm.Perm.Status
		default:
			//todo: make into error
			log.Println("[error] bad db state")
		}
	}

	_ = tx.Commit()

	return &res, nil
}

func (sqlite *SQLiteDB) GetRoles(pageSize int, offset int) ([]string, error) {
	var role string
	var roles []string
	rows, err := sqlite.db.Query("SELECT roleid FROM Roles WHERE roleTypeRK=0 LIMIT ? OFFSET ?", pageSize, offset)
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

func (sqlite *SQLiteDB) GiveRole(keyid string, roles ...string) error {
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

func (sqlite *SQLiteDB) TakeRole(keyid string, roles ...string) error {
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

func (sqlite *SQLiteDB) GrantPermission(permission *entities.Permission, roles ...string) error {
	//begin transaction
	tx, err := sqlite.db.Begin()
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}

	//get permissionid of existing/new permission node
	permissionID, err := sqlite.constructPermNode(tx, permission)
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}

	//add roles to permission node
	for _, role := range roles {
		err = sqlite.grantPermNode(tx, permissionID, role)
		if err != nil {
			rollbackOrPanic(tx)
			return err
		}
	}

	commitOrPanic(tx)

	return nil
}

/*
RevokePermission Removes a permission from the specified roles
-Retrieves a permissionID for the given resourceID, operationType, and denyAllow status
-Removes the specified role(s) from the RolePermIntersect
-Checks if any roles are still associated with that permission node
-Removes the permission node by id if it has no associated roles
*/
func (sqlite *SQLiteDB) RevokePermission(permission *entities.Permission, roles ...string) error {
	tx, e := sqlite.db.Begin()
	if e != nil {
		return e
	}

	var permissionID string

	//get permissionID from database
	row := tx.QueryRow("SELECT permissionid FROM Permissions WHERE resourceid = ? AND permTypeRWMD = ? AND permTypeDenyAllow = ?", permission.ResourceID, permission.TypeRWMD, permission.Status)
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

func (sqlite *SQLiteDB) SetRateLimit(key *entities.Key, limit *entities.RateLimit) error {
	//TODO implement me
	panic("implement me")
}

/*               *\
	 Object
	 Deletion
\*               */

func (sqlite *SQLiteDB) DeleteRole(name string) error {
	//TODO implement me
	panic("implement me")
}

func (sqlite *SQLiteDB) DeleteKey(id string) error {
	//TODO implement me
	panic("implement me")
}

func (sqlite *SQLiteDB) DeleteResource(id string) error {
	//TODO implement me
	panic("implement me")
}

func (sqlite *SQLiteDB) GetRateLimitData(ratelimitid string) (*entities.RateLimit, error) {
	row := sqlite.qm.GetRateLimitDataByID.QueryRow(ratelimitid)
	var rateLimit entities.RateLimit
	var reset int64
	err := row.Scan(&rateLimit.Limit, &reset)
	if err != nil {
		return nil, err
	}
	rateLimit.ID = ratelimitid
	rateLimit.Reset = serde.Duration(reset * int64(time.Millisecond))

	return &rateLimit, nil
}

func (sqlite *SQLiteDB) GetKeyRateLimitID(keyID string) (string, error) {
	var rateLimitID sql.NullString
	row := sqlite.qm.GetKeyRateLimitID.QueryRow(keyID)
	err := row.Scan(&rateLimitID)
	if err != nil {
		row = sqlite.qm.GetKeyIDIfExists.QueryRow(keyID)
		err = row.Scan(&keyID)
		if err == sql.ErrNoRows {
			return keyID, dberr.ErrKeyMissing
		} else {
			return keyID, err
		}
	}
	if rateLimitID.Valid {
		return rateLimitID.String, nil
	} else {
		return "", nil
	}
}

func (sqlite *SQLiteDB) UpdateRateLimit(rateLimitID string, rateLimit *entities.RateLimit) error {
	tx, err := sqlite.db.Begin()
	if err != nil {
		return err
	}
	stmt := tx.Stmt(sqlite.qm.UpdRateLimitData)
	res, err := stmt.Exec(rateLimitID, rateLimit.ID, rateLimit.Limit, rateLimit.Reset)
	if err != nil {
		return err
	}
	rowNum, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowNum != 1 {
		return errors.New(fmt.Sprintf("Failed to update rateLimit %d rows affected", rowNum))
	}
	return nil
}

func (sqlite *SQLiteDB) DeleteRateLimit(rateLimitID string) error {
	tx, err := sqlite.db.Begin()
	if err != nil {
		return err
	}
	stmt := tx.Stmt(sqlite.qm.UpdRateLimitData)
	_, err = stmt.Exec(rateLimitID)
	if err != nil {
		return err
	}
	return nil
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

func (sqlite *SQLiteDB) createResourcePermission(tx *sql.Tx, resource *entities.Resource, permMap map[string]bool, operationType types.OperationType) error {
	var err error
	var allowID int64 = -1
	var denyID int64 = -1

	for role, status := range permMap {
		if status && allowID == -1 {
			allowID, err = sqlite.constructPermNode(tx, &entities.Permission{ResourceID: resource.ID, TypeRWMD: operationType, Status: status})
			if err != nil {
				return err
			}
			err = sqlite.grantPermNode(tx, allowID, role)
			if err != nil {
				return err
			}
		} else if denyID == -1 {
			denyID, err = sqlite.constructPermNode(tx, &entities.Permission{ResourceID: resource.ID, TypeRWMD: operationType, Status: status})
			if err != nil {
				return err
			}
			err = sqlite.grantPermNode(tx, denyID, role)
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

// get permissionID of existing node or construct new permission node
func (sqlite *SQLiteDB) constructPermNode(tx *sql.Tx, permission *entities.Permission) (permissionID int64, err error) {
	stmt := tx.Stmt(sqlite.qm.GetPermissionIDByData)
	row := stmt.QueryRow(permission.ResourceID, permission.TypeRWMD, permission.Status)
	err = row.Scan(&permissionID)
	if err == sql.ErrNoRows {
		stmt = tx.Stmt(sqlite.qm.InsPermissionData)
		res, err := stmt.Exec(permission.ResourceID, permission.TypeRWMD, permission.Status)
		if err != nil {
			rollbackOrPanic(tx)
			return -1, err
		}
		permissionID, err = res.LastInsertId()
	} else if err != nil {
		rollbackOrPanic(tx)
		return -1, err
	}
	return permissionID, nil
}

// add roles to permission node
//
//todo:consider using in GrantPermission
func (sqlite *SQLiteDB) grantPermNode(tx *sql.Tx, permissionID int64, role string) error {
	stmt := tx.Stmt(sqlite.qm.InsRolePermIntersectData)
	_, err := stmt.Exec(role, permissionID)
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}
	return nil
}
