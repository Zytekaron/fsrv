package sqlite

import (
	"database/sql"
	"errors"
	"os"
)
import _ "github.com/mattn/go-sqlite3"

type SQLiteDB struct {
	db *sql.DB
}

func New(databaseFile string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		return nil, err
	}
	return &SQLiteDB{db}, nil
}

func Exists(databaseFile string) (bool, error) {
	_, err := os.Stat(databaseFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		} else {
			return false, err
		}
	}

	f, err := os.OpenFile(databaseFile, os.O_RDWR, 666)
	defer f.Close()
	if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func Create(databaseFile string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		return nil, err
	}

	_, err = db.Query(`
		CREATE TABLE IF NOT EXISTS Ratelimits (
		    ratelimitID TEXT PRIMARY KEY ,
		    requests INTEGER NOT NULL,
		    reset INTEGER NOT NULL
		);



		CREATE TABLE IF NOT EXISTS Keys (
			id TEXT PRIMARY KEY,
			comment TEXT,
			ratelimitID TEXT NOT NULL,
			expires INTEGER NOT NULL, --unix millis
			created INTEGER NOT NULL, --unix millis
					
			FOREIGN KEY (ratelimitID) REFERENCES Ratelimits(ratelimitID)
		);



		CREATE TABLE IF NOT EXISTS Roles(
			keyid TEXT,
			role TEXT,
			
			FOREIGN KEY (keyid) REFERENCES Keys(id)
		);



		CREATE TABLE IF NOT EXISTS Permissions (
		    id INTEGER,
			keyid TEXT,
			type TEXT, --RK=read-key, WK=write-key, RR=read-roles, WR=write-roles
		                                       
		    FOREIGN KEY (keyid) REFERENCES Keys(id)
		);
	`)

	if err != nil {
		return nil, err
	}
	return &SQLiteDB{db}, nil
}
