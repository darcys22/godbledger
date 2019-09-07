package db

import (
	"strconv"
	"time"

	"github.com/darcys22/godbledger/server/core"

	_ "github.com/mattn/go-sqlite3"
)

func (db *LedgerDB) SafeAddUser(usr *core.User) error {
	u, _ := db.FindUser(usr.Name)
	if u != nil {
		return nil
	}
	return db.AddUser(usr)

}

func (db *LedgerDB) AddCurrency(usr *core.Currency) error {
	return nil
}

func (db *LedgerDB) AddAccount(usr *core.Account) error {
	return nil
}

func (db *LedgerDB) AddUser(usr *core.User) error {
	log.Info("Adding User to DB")
	insertUser := `
		INSERT INTO users(user_id, username)
			VALUES(?,?);
	`
	tx, _ := db.DB.Begin()
	stmt, _ := tx.Prepare(insertUser)
	log.Debug("Query: " + insertUser)
	res, err := stmt.Exec(usr.Id, usr.Name)
	if err != nil {
		log.Fatal(err)
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("ID = %d, affected = %d\n", lastId, rowCnt)

	tx.Commit()

	return err
}

func (db *LedgerDB) FindUser(pubKey string) (*core.User, error) {
	var resp core.User
	log.Info("Searching User in DB")
	err := db.DB.QueryRow(`SELECT * FROM users WHERE username = $1 LIMIT 1`, pubKey).Scan(&resp.Id, &resp.Name)
	if err != nil {
		return nil, err
	}
	return &resp, nil
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
