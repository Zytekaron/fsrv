package sqlite

import (
	"database/sql"
	"fsrv/src/database/entities"
)

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

/////////////////////////////////////////
//									   //
/*-------------------------------------*\
//       NON INTERFACE FUNCTIONS	   //
\*-------------------------------------*/
//									   //
/////////////////////////////////////////

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
