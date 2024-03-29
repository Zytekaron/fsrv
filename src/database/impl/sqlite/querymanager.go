package sqlite

import (
	"database/sql"
	_ "embed"
	"log"
	"reflect"
)

type QueryManager struct {
	InsKeyData                                   *sql.Stmt
	GetRateLimitIDIfExists                       *sql.Stmt
	InsRateLimitData                             *sql.Stmt
	GetRoleIDIfExists                            *sql.Stmt
	GetKeyIDIfExists                             *sql.Stmt
	InsKeyRoleIntersectData                      *sql.Stmt
	InsRoleData                                  *sql.Stmt
	InsResourceData                              *sql.Stmt
	GetKeyIDs                                    *sql.Stmt
	GetKeyData                                   *sql.Stmt
	GetRateLimitDataByID                         *sql.Stmt
	GetRolesByKeyIDOrdered                       *sql.Stmt
	GetResourceIDs                               *sql.Stmt
	GetResourceFlagsByID                         *sql.Stmt
	GetRoleIDs                                   *sql.Stmt
	GetPermissionIDByData                        *sql.Stmt
	InsPermissionData                            *sql.Stmt
	GetPKOfLastInserted                          *sql.Stmt
	GetRolePermIntersectReferencesByPermissionID *sql.Stmt
	InsRolePermIntersectData                     *sql.Stmt
	GetResourceRoles                             *sql.Stmt
	GetKeyRateLimitID                            *sql.Stmt
	UpdRateLimitData                             *sql.Stmt
	UpdKeyRateLimitID                            *sql.Stmt
	DelPermissionByID                            *sql.Stmt
	DelRateLimitByID                             *sql.Stmt
	DelKeyByID                                   *sql.Stmt
	DelResourceByID                              *sql.Stmt
	DelRoleByID                                  *sql.Stmt
	DelPermissionByResourceID                    *sql.Stmt
	DelRPIEntryByRoleID                          *sql.Stmt
}

//go:embed readqueries/getResourceRoles.sql
var sqliteGetResourceRoles string

func NewQueryManager(db *sql.DB) (qm *QueryManager, err error) {
	//create obj
	var queryManager QueryManager
	qm = &queryManager

	//Insert Operations
	qm.InsKeyData, err = db.Prepare("INSERT INTO Keys (keyid, note, ratelimitid, expires, created) VALUES (?, ?, ?, ?, ?)") //CreateKey
	if err != nil {
		return qm, err
	}
	qm.InsRateLimitData, err = db.Prepare("INSERT INTO Ratelimits (ratelimitid, requests, burst, reset) VALUES (?, ?, ?, ?)") //CreateKey
	if err != nil {
		return qm, err
	}
	qm.InsKeyRoleIntersectData, err = db.Prepare("INSERT INTO KeyRoleIntersect (keyid, roleid) VALUES (?, ?)") //CreateKey
	if err != nil {
		return qm, err
	}
	qm.InsRoleData, err = db.Prepare("INSERT INTO Roles (roleid, roleTypeRK, rolePrecedence) VALUES (?, ?, ?)") //CreateKey, CreateRole
	if err != nil {
		return qm, err
	}
	qm.InsResourceData, err = db.Prepare("INSERT INTO Resources (resourceid,flags) VALUES (?, ?)") //CreateResource
	if err != nil {
		return qm, err
	}
	qm.InsPermissionData, err = db.Prepare("INSERT INTO Permissions (resourceid, permTypeRWMD, permTypeDenyAllow) VALUES (?, ?, ?)") //GrantPermission
	if err != nil {
		return qm, err
	}

	//Get operations
	qm.GetRateLimitIDIfExists, err = db.Prepare("SELECT ratelimitid FROM main.Ratelimits WHERE ratelimitid = ?") //CreateKey
	if err != nil {
		return qm, err
	}
	qm.GetRoleIDIfExists, err = db.Prepare("SELECT roleid FROM Roles WHERE roleid = ?") //CreateKey
	if err != nil {
		return qm, err
	}
	qm.GetKeyIDIfExists, err = db.Prepare("SELECT keyid FROM Keys WHERE keyid = ?")
	if err != nil {
		return qm, err
	}
	qm.GetKeyIDs, err = db.Prepare("SELECT keyid FROM Keys LIMIT ? OFFSET ?") //GetKeyIDs
	if err != nil {
		return qm, err
	}
	qm.GetKeyData, err = db.Prepare("SELECT note, ratelimitid, created, expires FROM Keys WHERE keyid = ?") //GetKeyData
	if err != nil {
		return qm, err
	}
	qm.GetRateLimitDataByID, err = db.Prepare("SELECT requests, burst, reset FROM Ratelimits WHERE ratelimitid = ?") //GetKeyData
	if err != nil {
		return qm, err
	}
	qm.GetRolesByKeyIDOrdered, err = db.Prepare("SELECT Roles.roleid FROM Roles JOIN KeyRoleIntersect KRI on Roles.roleid = KRI.roleid WHERE keyid = ? ORDER BY rolePrecedence") //GetKeyData
	if err != nil {
		return qm, err
	}
	qm.GetResourceIDs, err = db.Prepare("SELECT resourceid FROM Resources LIMIT ? OFFSET ?") //GetResourceIDs
	if err != nil {
		return qm, err
	}
	qm.GetResourceFlagsByID, err = db.Prepare("SELECT flags FROM Resources WHERE resourceid = ?") //GetResourceData
	if err != nil {
		return qm, err
	}
	qm.GetRoleIDs, err = db.Prepare("SELECT roleid FROM Roles WHERE roleTypeRK=0 LIMIT ? OFFSET ?") //GetRoles
	if err != nil {
		return qm, err
	}
	qm.GetPermissionIDByData, err = db.Prepare("SELECT permissionid FROM Permissions WHERE resourceid = ? AND permTypeRWMD = ? AND permTypeDenyAllow = ?") //GrantPermission, RevokePermission
	if err != nil {
		return qm, err
	}
	qm.GetPKOfLastInserted, err = db.Prepare("SELECT last_insert_rowid()") //GrantPermission
	if err != nil {
		return qm, err
	}
	qm.GetRolePermIntersectReferencesByPermissionID, err = db.Prepare("SELECT COUNT(permissionid) FROM RolePermIntersect WHERE permissionid = ?") //RevokePermission
	if err != nil {
		return qm, err
	}
	qm.InsRolePermIntersectData, err = db.Prepare("INSERT INTO RolePermIntersect (roleid, permissionid) VALUES (?,?)")
	if err != nil {
		return qm, err
	}
	qm.GetResourceRoles, err = db.Prepare(sqliteGetResourceRoles)
	if err != nil {
		return qm, err
	}
	qm.GetKeyRateLimitID, err = db.Prepare("SELECT ratelimitid FROM Keys WHERE keyid=?")
	if err != nil {
		return qm, err
	}
	qm.UpdRateLimitData, err = db.Prepare("UPDATE Ratelimits SET ratelimitid = ?, requests = ?, burst = ?, reset = ? WHERE ratelimitid = ?")
	if err != nil {
		return qm, err
	}
	qm.UpdKeyRateLimitID, err = db.Prepare("UPDATE Keys SET ratelimitid = ? WHERE keyid = ?")
	if err != nil {
		return qm, err
	}

	//Delete operations
	qm.DelPermissionByID, err = db.Prepare("DELETE FROM Permissions WHERE permissionid = ?") //RevokePermission
	if err != nil {
		return qm, err
	}
	qm.DelRateLimitByID, err = db.Prepare("DELETE FROM Ratelimits WHERE ratelimitid = ?") //DeleteRateLimit
	if err != nil {
		return qm, err
	}
	qm.DelKeyByID, err = db.Prepare("DELETE FROM Keys WHERE keyid = ?") //DeleteKey
	if err != nil {
		return qm, err
	}
	qm.DelResourceByID, err = db.Prepare("DELETE FROM Resources WHERE resourceid = ?")
	if err != nil {
		return qm, err
	}
	qm.DelRoleByID, err = db.Prepare("DELETE FROM Roles WHERE roleid = ?")
	if err != nil {
		return qm, err
	}
	qm.DelPermissionByResourceID, err = db.Prepare("DELETE FROM Permissions WHERE resourceid = ?")
	if err != nil {
		return qm, err
	}
	qm.DelRPIEntryByRoleID, err = db.Prepare("DELETE FROM RolePermIntersect WHERE roleid = ?")
	if err != nil {
		return qm, err
	}

	//qm.q, err = db.Prepare("")
	//if err != nil {
	//	return qm, err
	//}

	return qm, nil
}

// todo: test
func (qm *QueryManager) freePreparedQueries() error {
	v := reflect.ValueOf(qm).Elem()
	count := v.NumField()

	for i := 0; i < count; i++ {
		vi := v.Index(i)
		if vi.Type().Name() == "*sql.Stmt" {
			err := vi.Interface().(*sql.Stmt).Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// todo: test
func (qm *QueryManager) prepVarLenQuery(db *sql.DB, baseQuery string, repeatedSection string, conclusion string, repeats int) ([]*sql.Stmt, error) {
	if repeats < 1 {
		log.Panicf("prepareVariableLengthQuery: bad argument value for repeats (%d is < 1)", repeats)
	}

	stmts := make([]*sql.Stmt, repeats)
	var err error
	for i := 0; i < repeats; i++ {
		baseQuery += repeatedSection
		baseQuery += conclusion
		stmts[i], err = db.Prepare(baseQuery)
		if err != nil {
			return nil, err
		}
		baseQuery = baseQuery[:len(baseQuery)-len(conclusion)]
	}

	return stmts, nil
}
