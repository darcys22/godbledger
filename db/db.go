package db

import (
	"database/sql"
	"os"
	//"path"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	_ "github.com/mattn/go-sqlite3"
)

var log = logrus.WithField("prefix", "ledgerdb")

type LedgerDB struct {
	DB           *sql.DB
	DatabasePath string
}

// Close closes the underlying database.
func (db *LedgerDB) Close() error {
	return db.DB.Close()
}

// NewDB initializes a new DB. If the genesis block and states do not exist, this method creates it.
func NewDB(dirPath string) (*LedgerDB, error) {
	log.Info("Creating DB")
	//if err := os.MkdirAll(dirPath, 0700); err != nil {
	//return nil, err
	//}
	//datafile := path.Join(dirPath, "ledger.db")
	//SqliteDB, err := sql.Open("sqlite3", datafile)
	SqliteDB, err := sql.Open("sqlite3", "ledger.db")
	if err != nil {
		return nil, err
	}

	db := &LedgerDB{DB: SqliteDB, DatabasePath: dirPath}

	return db, err

}

// ClearDB removes the previously stored directory at the data directory.
func ClearDB(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(dirPath)
}

func (db *LedgerDB) TestDB() error {
	log.Info("Testing DB")
	createDB := "create table if not exists pages (title text, body blob, timestamp text)"
	log.Debug("Query: " + createDB)
	res, err := db.DB.Exec(createDB)

	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("ID = %d, affected = %d\n", lastId, rowCnt)

	tx, _ := db.DB.Begin()

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	stmt, _ := tx.Prepare("insert into pages (title, body, timestamp) values (?, ?, ?)")
	log.Debug("Query: Insert")
	res, err = stmt.Exec("Sean", "Body", timestamp)

	lastId, err = res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	rowCnt, err = res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("ID = %d, affected = %d\n", lastId, rowCnt)

	tx.Commit()

	return err
}
