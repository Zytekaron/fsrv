package sqlite

import (
	"database/sql"
	"errors"
	"fsrv/src/database"
	"fsrv/src/database/entities"
	"fsrv/utils/serde"
	"time"
)

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

func (sqlite *SQLiteDB) DeleteKey(id string) error {
	//TODO implement me
	panic("implement me")
}

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

func (sqlite *SQLiteDB) GetKeyRateLimitID(keyID string) (string, error) {
	var rateLimitID sql.NullString
	row := sqlite.qm.GetKeyRateLimitID.QueryRow(keyID)
	err := row.Scan(&rateLimitID)
	if err != nil {
		row = sqlite.qm.GetKeyIDIfExists.QueryRow(keyID)
		err = row.Scan(&keyID)
		if err == sql.ErrNoRows {
			return keyID, database.ErrKeyMissing
		} else {
			return "", err
		}
	}
	if rateLimitID.Valid {
		return rateLimitID.String, nil
	} else {
		return "", err
	}
}
