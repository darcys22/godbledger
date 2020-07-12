package mysqldb

import (
	"database/sql"

	"github.com/sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
)

var log = logrus.WithField("prefix", "MySQL")

type Database struct {
	DB               *sql.DB
	ConnectionString string
}

// Close closes the underlying database.
func (db *Database) Close() error {
	return db.DB.Close()
}

func DSN(DB_USER, DB_PASS, DB_HOST, DB_NAME string) string {
	return DB_USER + ":" + DB_PASS + "@" + DB_HOST + "/" + DB_NAME + "?charset=utf8&parseTime=true"
	//return DB_USER + ":" + DB_PASS + "@" + DB_HOST + "/"
}

func ValidateConnectionString(connection_string string) string {
	//if connection_string == "" {
	//DB_HOST := "tcp(127.0.0.1:3306)"
	//DB_NAME := "ledger"
	//DB_USER := "godbledger"
	//DB_PASS := "password"
	//connection_string = DSN(DB_USER, DB_PASS, DB_HOST, DB_NAME)
	//}
	return connection_string + "?charset=utf8&parseTime=true"
}

// NewDB initializes a new DB.
func NewDB(connection_string string) (*Database, error) {
	log.Debug(connection_string)
	MySQLDB, err := sql.Open("mysql", ValidateConnectionString(connection_string))
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}

	db := &Database{DB: MySQLDB, ConnectionString: connection_string}

	return db, nil
}

func (db *Database) InitDB() error {
	log.Info("Initialising DB Table")

	//USERS
	createDB := `
	CREATE TABLE IF NOT EXISTS users (
		user_id VARCHAR(255) NOT NULL,
		username VARCHAR(255) NOT NULL,
		PRIMARY KEY(user_id)
	);`
	log.Debug("Query: " + createDB)
	_, err := db.DB.Exec(createDB)
	if err != nil {
		log.Fatalf("Creating users table failed: %s", err)
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
		log.Fatalf("Creating accounts table failed: %s", err)
	}

	//TAGS
	createDB = `
	CREATE TABLE IF NOT EXISTS tags (
		tag_id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
		tag_name VARCHAR(100) NOT NULL UNIQUE
	);`
	log.Debug("Query: " + createDB)
	_, err = db.DB.Exec(createDB)
	if err != nil {
		log.Fatalf("Creating tags table failed: %s", err)
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
		log.Fatalf("Creating Account_Tag table failed: %s", err)
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
		poster_user_id VARCHAR(255),
		PRIMARY KEY(transaction_id),
    FOREIGN KEY (poster_user_id) REFERENCES users (user_id) ON DELETE RESTRICT
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

	//TAGS FOR Transactions
	createDB = `
	CREATE TABLE IF NOT EXISTS transaction_tag (
    transaction_id VARCHAR(255) NOT NULL,
    tag_id INTEGER NOT NULL,
    FOREIGN KEY (transaction_id) REFERENCES transactions (transaction_id) ON DELETE RESTRICT ON UPDATE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags (tag_id) ON DELETE RESTRICT ON UPDATE CASCADE,
    PRIMARY KEY (transaction_id, tag_id)
	);`
	log.Debug("Query: " + createDB)
	_, err = db.DB.Exec(createDB)
	if err != nil {
		log.Fatalf("Creating Transaction_Tag table failed: %s", err)
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
	return err
}

// ClearDB drops all tables
func (db *Database) ClearDB() error {

	//DROP TABLES
	dropDB := `
				SET foreign_key_checks = 0;
				SELECT
					 'DROP TABLE IF EXISTS ' + table_name + ';'
				FROM
						information_schema.tables
				WHERE
						table_schema = "ledger";
				SET foreign_key_checks = 0;
			`
	log.Debug("Query: " + dropDB)
	_, err := db.DB.Exec(dropDB)
	if err != nil {
		log.Fatalf("Dropping table failed with: %s", err)
		return err
	}
	return nil
}
