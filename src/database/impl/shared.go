package dbimpl

import (
	"errors"
	"os"
)

// Exists checks if a database file exists, is readable, and is writable
func Exists(databaseFile string) error {
	_, err := os.Stat(databaseFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		} else {
			return err
		}
	}

	f, err := os.OpenFile(databaseFile, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}
