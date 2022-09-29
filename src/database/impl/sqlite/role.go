package sqlite

import (
	"database/sql"
	"fsrv/src/database/entities"
)

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

func (sqlite *SQLiteDB) DeleteRole(name string) error {
	//begin transaction
	tx, err := sqlite.db.Begin()
	if err != nil {
		return err
	}

	//delete associated RolePermIntersect entries
	stmt := tx.Stmt(sqlite.qm.DelRPIEntryByRoleID)
	_, err = stmt.Exec(name)
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}

	//delete underlying role
	stmt = tx.Stmt(sqlite.qm.DelRoleByID)
	_, err = stmt.Exec(name)
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}

	//commit
	commitOrPanic(tx)
	return nil
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
