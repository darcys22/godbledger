package mysqldb

import (
	"database/sql"
	"errors"
	"regexp"

	"github.com/sirupsen/logrus"

	"github.com/go-sql-driver/mysql"
)

var log = logrus.WithField("prefix", "MySQL")
var dsnRegex = regexp.MustCompile(`\:(.+?)\@`)

type Database struct {
	DB               *sql.DB
	ConnectionString string
}

// Close closes the underlying database.
func (db *Database) Close() error {
	return db.DB.Close()
}

func ValidateConnectionString(dsn string) (string, error) {

	if dsn == "" {
		return "", errors.New("Connection string not provided")
	}

	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		log.Warnf("Connection string could not be parsed: %s", err.Error())
		return "", err
	}
	log.Debugf("DB_ADDR := %s", cfg.Addr)
	log.Debugf("DB_NET := %s", cfg.Net)
	log.Debugf("DB_DBNAME := %s", cfg.DBName)
	log.Debugf("DB_USER := %s", cfg.User)
	log.Debugf("PARAMS := %v", cfg.Params)
	if !cfg.ParseTime {
		cfg.ParseTime = true
	}
	charset, ok := cfg.Params["charset"]
	if !(ok && charset == "utf8") {
		if cfg.Params == nil {
			cfg.Params = make(map[string]string)
		}
		cfg.Params["charset"] = "utf8"
	}

	log.Debugf("ParseTime := %v", cfg.ParseTime)
	log.Debugf("Charset := %s", cfg.Params["charset"])

	dsnString := cfg.FormatDSN()
	log.Debugf("DSN := %s", redactPassword(dsnString))

	return dsnString, nil
}

func redactPassword(rawDSNString string) string {
	cleanedDSNString := dsnRegex.ReplaceAll([]byte(rawDSNString), []byte(":**REDACTED**@"))
	return string(cleanedDSNString)
}

// NewDB initializes a new DB.
func NewDB(connection_string string) (*Database, error) {
	validatedString, err := ValidateConnectionString(connection_string)
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}
	MySQLDB, err := sql.Open("mysql", validatedString)
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

	//RECONCILIATIONS
	createDB = `
	CREATE TABLE IF NOT EXISTS reconciliations (
		reconciliation_id VARCHAR(255) NOT NULL,
		split_id VARCHAR(255) NOT NULL,
		FOREIGN KEY (split_id) REFERENCES splits (split_id) ON DELETE RESTRICT ON UPDATE CASCADE,
		PRIMARY KEY (reconciliation_id, split_id)
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
				DROP DATABASE ledger;
			`
	log.Debug("Query: " + dropDB)
	_, err := db.DB.Exec(dropDB)
	if err != nil {
		log.Fatalf("Dropping table failed with: %s", err)
		return err
	}

	//CREATE NEW DATABASE
	newDB := `
				CREATE DATABASE ledger;
			`
	log.Debug("Query: " + newDB)
	_, err = db.DB.Exec(newDB)
	if err != nil {
		log.Fatalf("Creating table failed with: %s", err)
		return err
	}

	//USE NEW DATABASE
	newDB = `
				USE ledger;
			`
	log.Debug("Query: " + newDB)
	_, err = db.DB.Exec(newDB)
	if err != nil {
		log.Fatalf("Creating table failed with: %s", err)
		return err
	}
	return nil
}
