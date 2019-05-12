package db

import (
	"database/sql"
	"errors"
	"os"
	"path"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	_ "github.com/mattn/go-sqlite3"
)

var log = logrus.WithField("prefix", "ledgerdb")

type LedgerDB struct {
	db           *sql.DB
	DatabasePath string
}

// Close closes the underlying database.
func (db *LedgerDB) Close() error {
	return db.db.Close()
}

// NewDB initializes a new DB. If the genesis block and states do not exist, this method creates it.
func NewDB(dirPath string) (*LedgerDB, error) {
	if err := os.MkdirAll(dirPath, 0700); err != nil {
		return nil, err
	}
	datafile := path.Join(dirPath, "ledger.db")
	SqliteDB, err := sql.Open("sqlite3", datafile)
	if err != nil {
		return nil, err
	}

	db := &LedgerDB{db: SqliteDB, DatabasePath: dirPath}

	return db, err

}

// ClearDB removes the previously stored directory at the data directory.
func ClearDB(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(dirPath)
}
