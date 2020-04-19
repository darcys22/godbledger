package mysqldb

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/darcys22/godbledger/godbledger/core"

	_ "github.com/go-sql-driver/mysql"
)

func (db *Database) AddTransaction(txn *core.Transaction) error {
	log.Info("Adding Transaction to DB")
	insertTransaction := `
		INSERT INTO transactions(transaction_id, postdate, brief)
			VALUES(?,?,?);
	`
	tx, _ := db.DB.Begin()
	stmt, _ := tx.Prepare(insertTransaction)
	log.Debug("Query: " + insertTransaction)
	res, err := stmt.Exec(txn.Id, txn.Postdate, string(txn.Description[:]))
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

	sqlStr := "INSERT INTO splits(transaction_id, split_id, split_date, description, currency, amount) VALUES "
	vals := []interface{}{}
	sqlAccStr := "INSERT INTO split_accounts(split_id, account_id) VALUES "
	accVals := []interface{}{}

	for _, split := range txn.Splits {
		sqlStr += "(?, ?, ?, ?, ?, ?),"
		vals = append(vals, txn.Id, split.Id, split.Date, string(split.Description[:]), split.Currency.Name, split.Amount.Int64())
		for _, acc := range split.Accounts {
			sqlAccStr += "(?, ?),"
			accVals = append(accVals, split.Id, acc.Code)
		}
	}

	sqlStr = strings.TrimSuffix(sqlStr, ",")
	tx, _ = db.DB.Begin()
	stmt, _ = tx.Prepare(sqlStr)
	log.Debug("Query: " + sqlStr)
	log.Debugf("NumberVals = %d", len(vals))
	log.Info("Adding Split to DB")
	res, err = stmt.Exec(vals...)
	if err != nil {
		log.Fatal(err)
	}

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

	sqlAccStr = strings.TrimSuffix(sqlAccStr, ",")
	tx2, _ := db.DB.Begin()
	accStmt, _ := tx2.Prepare(sqlAccStr)
	log.Debug("Query: " + sqlAccStr)
	log.Info("Adding Split Accounts to DB")
	res, err = accStmt.Exec(accVals...)
	if err != nil {
		log.Fatal(err)
	}

	lastId, err = res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	rowCnt, err = res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("ID = %d, affected = %d\n", lastId, rowCnt)

	tx2.Commit()

	return err
}

func (db *Database) DeleteTransaction(txnID string) error {

	sqlStatement := `
	DELETE FROM transactions
	WHERE transaction_id = ?;`
	_, err := db.DB.Exec(sqlStatement, txnID)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) FindTag(tag string) (int, error) {
	var resp int
	log.Info("Searching Tag in DB")
	err := db.DB.QueryRow(`SELECT tag_id FROM tags WHERE tag_name = ? LIMIT 1`, tag).Scan(&resp)
	if err != nil {
		log.Debug("Find Tag Failed: ", err)
		return 0, err
	}
	return resp, nil
}

func (db *Database) AddTag(tag string) error {
	log.Info("Adding Tag to DB")
	insertTag := `
		INSERT INTO tags(tag_name)
			VALUES(?);
	`
	tx, _ := db.DB.Begin()
	stmt, _ := tx.Prepare(insertTag)
	log.Debug("Query: " + insertTag)
	res, err := stmt.Exec(tag)
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

	return nil
}

func (db *Database) SafeAddTag(tag string) error {
	u, _ := db.FindTag(strings.TrimSpace(tag))
	//if err != nil {
	//log.Debug(err)
	//return err
	//}
	if u != 0 {
		return nil
	}
	return db.AddTag(strings.TrimSpace(tag))
}

func (db *Database) SafeAddTagToAccount(account, tag string) error {
	err := db.SafeAddTag(tag)
	if err != nil {
		log.Debug(err)
		return err
	}
	tagID, _ := db.FindTag(tag)

	var accountID string
	err = db.DB.QueryRow(`SELECT account_id FROM accounts WHERE name = ? LIMIT 1`, account).Scan(&accountID)
	if err != nil {
		log.Debug(err)
		return err
	}

	return db.AddTagToAccount(accountID, tagID)
}

func (db *Database) AddTagToAccount(accountID string, tag int) error {
	var exists int
	err := db.DB.QueryRow(`SELECT EXISTS(SELECT * FROM account_tag where (account_id = ?) AND (tag_id = ?));`, accountID, tag).Scan(&exists)
	if err != nil {
		log.Debug(err)
		return err
	}
	if exists == 1 {
		return nil
	}

	insertTag := `
		INSERT INTO account_tag(account_id, tag_id)
			VALUES(?,?);
	`
	tx, _ := db.DB.Begin()
	stmt, _ := tx.Prepare(insertTag)
	log.Debug("Query: " + insertTag)
	res, err := stmt.Exec(accountID, tag)
	if err != nil {
		log.Debug(err)
		return err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		log.Debug(err)
		return err
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Debug(err)
		return err
	}
	log.Debugf("ID = %d, affected = %d\n", lastId, rowCnt)

	tx.Commit()

	return nil

}

func (db *Database) DeleteTagFromAccount(account, tag string) error {

	tagID, err := db.FindTag(tag)
	if err != nil {
		return err
	}

	sqlStatement := `
	DELETE FROM account_tag
	WHERE 
		tag_id = ?
	AND
		account_id = ?
	;`
	_, err = db.DB.Exec(sqlStatement, tagID, account)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) FindCurrency(cur string) (*core.Currency, error) {
	var resp core.Currency
	log.Info("Searching Currency in DB: ", cur)
	err := db.DB.QueryRow(`SELECT * FROM currencies WHERE name = ? LIMIT 1`, cur).Scan(&resp.Name, &resp.Decimals)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (db *Database) AddCurrency(cur *core.Currency) error {
	log.Info("Adding Currency to DB")
	insertCurrency := `
		INSERT INTO currencies(name,decimals)
			VALUES(?,?);
	`
	tx, _ := db.DB.Begin()
	stmt, _ := tx.Prepare(insertCurrency)
	log.Debug("Query: " + insertCurrency)
	res, err := stmt.Exec(cur.Name, cur.Decimals)
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

func (db *Database) SafeAddCurrency(cur *core.Currency) error {
	u, _ := db.FindCurrency(cur.Name)
	if u != nil {
		return nil
	}
	return db.AddCurrency(cur)
}

func (db *Database) FindAccount(code string) (*core.Account, error) {
	var resp core.Account
	log.Info("Searching Account in DB")
	err := db.DB.QueryRow(`SELECT * FROM accounts WHERE account_id = ? LIMIT 1`, code).Scan(&resp.Code, &resp.Name)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (db *Database) AddAccount(acc *core.Account) error {
	log.Info("Adding Account to DB")
	insertAccount := `
		INSERT INTO accounts(account_id, name)
			VALUES(?,?);
	`
	tx, _ := db.DB.Begin()
	stmt, _ := tx.Prepare(insertAccount)
	log.Debug("Query: " + insertAccount)
	res, err := stmt.Exec(strings.TrimSpace(acc.Code), strings.TrimSpace(acc.Name))
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

func (db *Database) SafeAddAccount(acc *core.Account) error {
	u, _ := db.FindAccount(acc.Code)
	if u != nil {
		return nil
	}
	return db.AddAccount(acc)

}

func (db *Database) FindUser(pubKey string) (*core.User, error) {
	var resp core.User
	log.Info("Searching User in DB")
	err := db.DB.QueryRow(`SELECT * FROM users WHERE username = ? LIMIT 1`, pubKey).Scan(&resp.Id, &resp.Name)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (db *Database) AddUser(usr *core.User) error {
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

func (db *Database) SafeAddUser(usr *core.User) error {
	u, _ := db.FindUser(usr.Name)
	if u != nil {
		return nil
	}
	return db.AddUser(usr)

}

func (db *Database) TestDB() error {
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

func (db *Database) GetTB(date time.Time) error {

	queryDB := `
			SELECT
					tags.tag_name,
					Table_Aggregate.account_id,
					sums
			FROM account_tag
					join ((SELECT
							split_accounts.account_id as account_id,
							SUM(splits.amount) as sums
						FROM splits 
							JOIN split_accounts 
							ON splits.split_id = split_accounts.split_id
						GROUP  BY split_accounts.account_id
							
						)) AS Table_Aggregate
						on account_tag.account_id = Table_Aggregate.account_id
					join tags
						on tags.tag_id = account_tag.tag_id
			order BY tags.tag_name
		;`

	rows, err := db.DB.Query(queryDB)
	if err != nil {
		log.Debug(err)
		return err
	}
	defer rows.Close()

	accounts := make(map[string][]*core.PDFAccount)
	totals := make(map[string]int)

	for rows.Next() {
		var t *core.PDFAccount
		var name string
		if err := rows.Scan(&name, &t.Account, &t.Amount); err != nil {
			log.Fatal(err)
		}
		log.Debugf("%v", t)
		if val, ok := accounts[name]; ok {
			accounts[name] = append(val, t)
			totals[name] = totals[name] + t.Amount
		} else {
			accounts[name] = []*core.PDFAccount{t}
			totals[name] = t.Amount
		}
	}
	if rows.Err() != nil {
		log.Fatal(err)
	}

	//for k, v := range accounts {
	//reporteroutput.Data = append(reporteroutput.Data, Tag{k, totals[k], v})
	//}

	return nil

}

func (db *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.Query(query, args...)

}
