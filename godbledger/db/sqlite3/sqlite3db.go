package sqlite3db

import (
	"database/sql"
	"os"
	"path"

	"github.com/sirupsen/logrus"

	_ "github.com/mattn/go-sqlite3"
)

var log = logrus.WithField("prefix", "SQLLite")

type Database struct {
	DB           *sql.DB
	DatabasePath string
}

// Close closes the underlying database.
func (db *Database) Close() error {
	return db.DB.Close()
}

// NewDB initializes a new DB.
func NewDB(dirPath string) (*Database, error) {
	log.Info("Creating DB")
	if err := os.MkdirAll(dirPath, 0700); err != nil {
		return nil, err
	}
	datafile := path.Join(dirPath, "ledger.db?_foreign_keys=true")
	SqliteDB, err := sql.Open("sqlite3", datafile)
	if err != nil {
		return nil, err
	}

	db := &Database{DB: SqliteDB, DatabasePath: dirPath}

	return db, err

}

func (db *Database) InitDB() error {
	log.Info("Initialising DB Table")

	//USERS
	createDB := `
	CREATE TABLE IF NOT EXISTS users (
		user_id INT NOT NULL,
		username VARCHAR(255) NOT NULL,
		PRIMARY KEY(user_id)
	);`
	log.Debug("Query: " + createDB)
	_, err := db.DB.Exec(createDB)
	if err != nil {
		log.Fatal(err)
	}

	//ACCOUNTS
	createDB = `
	CREATE TABLE IF NOT EXISTS accounts (
		account_id VARCHAR(255) NOT NULL,
		name VARCHAR(255) NOT NULL,
		PRIMARY KEY(account_id)
	);`
	log.Debug("Query: " + createDB)
	_, err = db.DB.Exec(createDB)
	if err != nil {
		log.Fatal(err)
	}

	//TAGS
	createDB = `
	CREATE TABLE IF NOT EXISTS tags (
		tag_id INTEGER PRIMARY KEY,
		tag_name VARCHAR(100) NOT NULL UNIQUE
	);`
	log.Debug("Query: " + createDB)
	_, err = db.DB.Exec(createDB)
	if err != nil {
		log.Fatal(err)
	}

	//TAGS FOR ACCOUNTS
	createDB = `
	CREATE TABLE IF NOT EXISTS account_tag (
    account_id VARCHAR(255) NOT NULL,
    tag_id INTEGER NOT NULL,
    FOREIGN KEY (account_id) REFERENCES accounts (account_id) ON DELETE RESTRICT ON UPDATE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags (tag_id) ON DELETE RESTRICT ON UPDATE CASCADE,
    PRIMARY KEY (account_id, tag_id)
	);`
	log.Debug("Query: " + createDB)
	_, err = db.DB.Exec(createDB)
	if err != nil {
		log.Fatal(err)
	}

	//CURRENCIES
	createDB = `
	CREATE TABLE IF NOT EXISTS currencies (
		name VARCHAR(255) NOT NULL,
		decimals INT NOT NULL,
		PRIMARY KEY(name)
	);`
	log.Debug("Query: " + createDB)
	_, err = db.DB.Exec(createDB)
	if err != nil {
		log.Fatal(err)
	}

	//TRANSACTIONS
	createDB = `
	CREATE TABLE IF NOT EXISTS transactions (
		transaction_id VARCHAR(255) NOT NULL,
		postdate DATETIME NOT NULL,
		brief VARCHAR(255),
		PRIMARY KEY(transaction_id)
	);`
	log.Debug("Query: " + createDB)
	_, err = db.DB.Exec(createDB)
	if err != nil {
		log.Fatal(err)
	}

	//TRANSACTIONS BODY
	createDB = `
	CREATE TABLE IF NOT EXISTS transactions_body (
		transaction_id VARCHAR(255) NOT NULL,
		body TEXT,
		FOREIGN KEY(transaction_id) REFERENCES transactions(transaction_id) ON DELETE CASCADE
	);`
	log.Debug("Query: " + createDB)
	_, err = db.DB.Exec(createDB)
	if err != nil {
		log.Fatal(err)
	}

	//LINE ITEMS FOR TRANSACTIONS (SPLITS)
	createDB = `
	CREATE TABLE IF NOT EXISTS splits (
		split_id VARCHAR(255) NOT NULL,
		split_date DATETIME,
		description VARCHAR(255),
		currency VARCHAR(255),
		amount BIGINT,
		transaction_id VARCHAR(255),
		FOREIGN KEY(transaction_id) REFERENCES transactions(transaction_id) ON DELETE CASCADE,
		PRIMARY KEY(split_id)
	);`
	log.Debug("Query: " + createDB)
	_, err = db.DB.Exec(createDB)
	if err != nil {
		log.Fatal(err)
	}

	//ACCOUNTS FOR SPLITS
	createDB = `
	CREATE TABLE IF NOT EXISTS split_accounts (
		id INT AUTO_INCREMENT PRIMARY KEY,
		split_id VARCHAR(255),
		account_id VARCHAR(255),
		FOREIGN KEY(split_id) REFERENCES splits(split_id) ON DELETE CASCADE,
		FOREIGN KEY(account_id) REFERENCES accounts(account_id) ON DELETE CASCADE
	);`
	log.Debug("Query: " + createDB)
	_, err = db.DB.Exec(createDB)
	if err != nil {
		log.Fatal(err)
	}

	//ENTITIES
	createDB = `
	CREATE TABLE IF NOT EXISTS entities (
		entity_id VARCHAR(255) NOT NULL,
		name VARCHAR(255) NOT NULL,
		tag VARCHAR(255),
		type VARCHAR(255),
		description VARCHAR(255),
		PRIMARY KEY(entity_id)
	);`
	log.Debug("Query: " + createDB)
	_, err = db.DB.Exec(createDB)
	if err != nil {
		log.Fatal(err)
	}

	//Default Currencies
	insertCurrency := `
		INSERT INTO currencies(name,decimals)
			VALUES("USD",2),
			("AUD",2),
			("GBP",2),
			("BTC",8),
			("ETH",9),
			("LOKI",9);
	`
	log.Debug("Query: " + insertCurrency)
	_, _ = db.DB.Exec(insertCurrency)
	return err
}

// ClearDB removes the previously stored directory at the data directory.
func ClearDB(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(dirPath)
}
