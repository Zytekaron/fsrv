package sqlite

import (
	"errors"
	"fmt"
	"fsrv/src/database/entities"
	"fsrv/utils/serde"
	"time"
)

func (sqlite *SQLiteDB) CreateRateLimit(limit *entities.RateLimit) error {
	_, err := sqlite.qm.InsRateLimitData.Exec(limit.ID, limit.Limit, limit.Burst, limit.Refill)
	return err
}

func (sqlite *SQLiteDB) DeleteRateLimit(rateLimitID string) error {
	return sqlite.deleteObjByID(sqlite.qm.DelRateLimitByID, rateLimitID)
}

func (sqlite *SQLiteDB) SetRateLimit(key *entities.Key, limitID string) error {
	tx, err := sqlite.db.Begin()
	if err != nil {
		return err
	}
	stmt := tx.Stmt(sqlite.qm.UpdKeyRateLimitID)
	_, err = stmt.Exec(key.ID, limitID)
	if err != nil {
		rollbackOrPanic(tx)
		return err
	}
	commitOrPanic(tx)
	return nil
}

// GetRateLimitData todo: wrap sql: error no rows
func (sqlite *SQLiteDB) GetRateLimitData(ratelimitid string) (*entities.RateLimit, error) {
	row := sqlite.qm.GetRateLimitDataByID.QueryRow(ratelimitid)
	var rateLimit entities.RateLimit
	var reset int64
	err := row.Scan(&rateLimit.Limit, &rateLimit.Burst, &reset)
	if err != nil {
		return nil, err
	}
	rateLimit.ID = ratelimitid
	rateLimit.Refill = serde.Duration(reset * int64(time.Millisecond))

	return &rateLimit, nil
}

func (sqlite *SQLiteDB) UpdateRateLimit(rateLimitID string, rateLimit *entities.RateLimit) error {
	tx, err := sqlite.db.Begin()
	if err != nil {
		return err
	}
	stmt := tx.Stmt(sqlite.qm.UpdRateLimitData)
	res, err := stmt.Exec(rateLimitID, rateLimit.ID, rateLimit.Limit, rateLimit.Burst, rateLimit.Refill)
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
