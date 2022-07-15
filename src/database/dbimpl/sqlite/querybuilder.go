package sqlite

import (
	"database/sql"
	"errors"
	"fsrv/src/database/entities"
	"fsrv/utils/serde"
	"time"
)

type KeyQueryBuilder struct {
	key entities.Key
	db  *sql.DB
}

func createKeyQueryBuilder(db SQLiteDB) *KeyQueryBuilder {
	return &KeyQueryBuilder{
		key: entities.Key{
			ID:               "",
			Comment:          "",
			RequestRateLimit: nil,
			Roles:            nil,
			ExpiresAt:        serde.Time(time.Now().AddDate(0, 1, 0)),
			CreatedAt:        serde.Time(time.Now()),
		},
		db: db.db,
	}
}

func (kqb *KeyQueryBuilder) withKey(keystring string) *KeyQueryBuilder {
	kqb.key.ID = keystring
	return kqb
}

func (kqb *KeyQueryBuilder) withRoles(str ...string) *KeyQueryBuilder {
	kqb.key.Roles = str
	return kqb
}

func (kqb *KeyQueryBuilder) withRequestRateLimit(rateLimit entities.RateLimit) *KeyQueryBuilder {
	kqb.key.RequestRateLimit = &rateLimit
	return kqb
}

func (kqb *KeyQueryBuilder) withExpiry(time time.Time) *KeyQueryBuilder {
	kqb.key.ExpiresAt = serde.Time(time)
	return kqb
}

func (kqb *KeyQueryBuilder) withCreationDate(time time.Time) *KeyQueryBuilder {
	kqb.key.CreatedAt = serde.Time(time)
	return kqb
}

func (kqb *KeyQueryBuilder) withComment(comment string) *KeyQueryBuilder {
	kqb.key.Comment = comment
	return kqb
}

func (kqb *KeyQueryBuilder) run() error {
	if kqb.key.ID != "" {
		return errors.New("required feild keyid not specified")
	}

	//begin transaction
	_, err := kqb.db.Query("BEGIN TRANSACTION;")
	if err != nil {
		//don't rollback the transaction starting
		return err
	}
	//create key record
	_, err = kqb.db.Query("INSERT INTO Keys (keyid, note, expires, created) VALUES (?, ?, ?, ?)",
		kqb.key.ID, kqb.key.Comment, time.Time(kqb.key.ExpiresAt).UnixMilli(), time.Time(kqb.key.CreatedAt).UnixMilli())
	if err != nil {
		_, _ = kqb.db.Query("ROLLBACK;")
		return err
	}

	//add RateLimit if exists
	if kqb.key.RequestRateLimit != nil {
		_, err = kqb.db.Query("INSERT INTO Ratelimits (keyid, requests, reset) VALUES (?, ?, ?)",
			kqb.key.ID, kqb.key.RequestRateLimit.Limit, time.Duration(kqb.key.RequestRateLimit.Reset).Milliseconds())
		if err != nil {
			_, _ = kqb.db.Query("ROLLBACK;")
			return err
		}
	}

	//add Roles
	var rows *sql.Rows
	var roleid int
	for _, role := range kqb.key.Roles {
		rows, err = kqb.db.Query("SELECT roleid FROM Roles WHERE roleName = ?", role)
		if err != nil {
			_, _ = kqb.db.Query("ROLLBACK;")
			return err
		}
		err = rows.Scan(roleid)
		if err != nil {
			_, _ = kqb.db.Query("ROLLBACK;")
			return err
		}

		_, err = kqb.db.Query("INSERT INTO KeyRoleIntersect (keyid, roleid) VALUES (?, ?)", kqb.key.ID, roleid)
		if err != nil {
			_, _ = kqb.db.Query("ROLLBACK;")
			return err
		}
	}

	//commit transaction results
	_, err = kqb.db.Query("COMMIT;")
	if err != nil {
		_, _ = kqb.db.Query("ROLLBACK;")
		return err
	}

	return nil
}
