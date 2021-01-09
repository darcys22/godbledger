package mysqldb

import (
	"database/sql"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/darcys22/godbledger/godbledger/core"

	_ "github.com/go-sql-driver/mysql"
)

func (db *Database) AddTransaction(txn *core.Transaction) (string, error) {
	log.Debug("Adding Transaction to DB")

	longDescription := false

	if len(txn.Description) > 255 {
		longDescription = true
	}

	posterID := ""
	err := db.DB.QueryRow(`SELECT user_id FROM users WHERE username = ? LIMIT 1`, txn.Poster.Name).Scan(&posterID)
	if err != nil {
		log.Fatal(err)
	}

	insertTransaction := `
		INSERT INTO transactions(transaction_id, postdate, brief,poster_user_id)
			VALUES(?,?,?,?);
	`
	tx, _ := db.DB.Begin()
	stmt, _ := tx.Prepare(insertTransaction)
	log.Debug("Query: " + insertTransaction)

	var res sql.Result
	if longDescription {
		res, err = stmt.Exec(txn.Id, txn.Postdate, string(txn.Description[:255]), posterID)
	} else {
		res, err = stmt.Exec(txn.Id, txn.Postdate, string(txn.Description[:]), posterID)
	}
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

	if longDescription {
		insertLongDescriptionTransaction := `
			INSERT INTO transactions_body(transaction_id, body)
				VALUES(?,?);
		`
		stmt, _ := tx.Prepare(insertLongDescriptionTransaction)
		log.Debug("Query: " + insertLongDescriptionTransaction)
		log.Debug("Txn Id: " + txn.Id)
		res, err := stmt.Exec(txn.Id, string(txn.Description[:]))
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
		log.Debug("Saving Long Description into extended table")
	}

	sqlStr := "INSERT INTO splits(transaction_id, split_id, split_date, description, currency, amount) VALUES "
	vals := []interface{}{}
	sqlAccStr := "INSERT INTO split_accounts(split_id, account_id) VALUES "
	accVals := []interface{}{}

	for _, split := range txn.Splits {
		sqlStr += "(?, ?, ?, ?, ?, ?),"
		//Todo:(sean) split is truncated at 255 bytes but should be handled better
		if len(split.Description) > 255 {
			vals = append(vals, txn.Id, split.Id, split.Date, string(split.Description[:255]), split.Currency.Name, split.Amount.Int64())
		} else {
			vals = append(vals, txn.Id, split.Id, split.Date, string(split.Description[:]), split.Currency.Name, split.Amount.Int64())
		}
		for _, acc := range split.Accounts {
			sqlAccStr += "(?, ?),"
			accVals = append(accVals, split.Id, strings.TrimSpace(acc.Code))
		}
	}

	sqlStr = strings.TrimSuffix(sqlStr, ",")
	stmt, _ = tx.Prepare(sqlStr)
	log.Debug("Query: " + sqlStr)
	log.Debugf("NumberVals = %d", len(vals))
	log.Debug("Adding Split to DB")
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

	sqlAccStr = strings.TrimSuffix(sqlAccStr, ",")
	accStmt, _ := tx.Prepare(sqlAccStr)
	log.Debug("Query: " + sqlAccStr)
	log.Debug("Adding Split Accounts to DB")
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

	tx.Commit()

	return txn.Id, err
}

func (db *Database) FindTransaction(txnID string) (*core.Transaction, error) {
	var resp core.Transaction
	var poster core.User
	log.Debug("Searching Transaction in DB: ", txnID)

	// Find the transaction body
	err := db.DB.QueryRow(`
			SELECT t.transaction_id,
						 t.postdate,
						 t.brief,
						 u.user_id,
						 u.username
			FROM   transactions AS t
						 JOIN users AS u
							 ON t.poster_user_id = u.user_id
			WHERE  t.transaction_id = ?
			LIMIT  1 
			`, txnID).Scan(&resp.Id, &resp.Postdate, &resp.Description, &poster.Id, &poster.Name)
	if err != nil {
		return nil, err
	}

	log.Debug("Searching Transaction splits in DB")

	// Find all splits relating to that transaction
	splits, err := db.Query(`
			SELECT s.split_id,
						 s.split_date,
						 s.description,
						 a.account_id,
						 a.NAME,
						 s.currency,
						 c.decimals,
						 s.amount
			FROM   splits AS s
						 JOIN split_accounts AS sa
							 ON s.split_id = sa.split_id
						 JOIN accounts AS a
							 ON sa.account_id = a. account_id
						 JOIN currencies AS c
							 ON s.currency = c.NAME
			WHERE  s.transaction_id = ?
			`, txnID)
	if err != nil {
		return nil, err
	}

	for splits.Next() {
		var split core.Split
		var account core.Account
		var cur core.Currency
		var amount int64
		// for each row, scan the result into our split object
		err = splits.Scan(&split.Id, &split.Date, &split.Description, &account.Code, &account.Name, &cur.Name, &cur.Decimals, &amount)
		if err != nil {
			return nil, err
		}
		split.Amount = big.NewInt(amount)
		split.Accounts = append(split.Accounts, &account)
		split.Currency = &cur
		resp.Splits = append(resp.Splits, &split)

	}

	return &resp, nil
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
	log.Debug("Searching Tag in DB")
	err := db.DB.QueryRow(`SELECT tag_id FROM tags WHERE tag_name = ? LIMIT 1`, tag).Scan(&resp)
	if err != nil {
		log.Debug("Find Tag Failed: ", err)
		return 0, err
	}
	return resp, nil
}

func (db *Database) AddTag(tag string) error {
	log.Debug("Adding Tag to DB")
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

func (db *Database) SafeAddTagToTransaction(txnID, tag string) error {
	err := db.SafeAddTag(tag)
	if err != nil {
		log.Debug(err)
		return err
	}
	tagID, _ := db.FindTag(tag)

	return db.AddTagToTransaction(txnID, tagID)
}

func (db *Database) AddTagToTransaction(txnID string, tag int) error {
	var exists int
	err := db.DB.QueryRow(`SELECT EXISTS(SELECT * FROM transaction_tag where (transaction_id = ?) AND (tag_id = ?));`, txnID, tag).Scan(&exists)
	if err != nil {
		log.Debug(err)
		return err
	}
	if exists == 1 {
		return nil
	}

	insertTag := `
		INSERT INTO transaction_tag(transaction_id, tag_id)
			VALUES(?,?);
	`
	tx, _ := db.DB.Begin()
	stmt, _ := tx.Prepare(insertTag)
	log.Debug("Query: " + insertTag)
	res, err := stmt.Exec(txnID, tag)
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

func (db *Database) DeleteTagFromTransaction(txnID, tag string) error {

	tagID, err := db.FindTag(tag)
	if err != nil {
		return err
	}

	sqlStatement := `
	DELETE FROM transaction_tag
	WHERE 
		tag_id = ?
	AND
		transaction_id = ?
	;`
	_, err = db.DB.Exec(sqlStatement, tagID, txnID)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) FindCurrency(cur string) (*core.Currency, error) {
	var resp core.Currency
	log.Debug("Searching Currency in DB: ", cur)
	err := db.DB.QueryRow(`SELECT * FROM currencies WHERE name = ? LIMIT 1`, strings.TrimSpace(cur)).Scan(&resp.Name, &resp.Decimals)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (db *Database) AddCurrency(cur *core.Currency) error {
	log.Debug("Adding Currency to DB")
	insertCurrency := `
		INSERT INTO currencies(name,decimals)
			VALUES(?,?);
	`
	tx, _ := db.DB.Begin()
	stmt, _ := tx.Prepare(insertCurrency)
	log.Debug("Query: " + insertCurrency)
	res, err := stmt.Exec(strings.TrimSpace(cur.Name), cur.Decimals)
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

func (db *Database) DeleteCurrency(currency string) error {

	sqlStatement := `
	DELETE FROM currencies
	WHERE name = ?;`
	_, err := db.DB.Exec(sqlStatement, currency)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) FindAccount(code string) (*core.Account, error) {
	var resp core.Account
	log.Debug("Searching Account in DB")
	err := db.DB.QueryRow(`SELECT * FROM accounts WHERE account_id = ? LIMIT 1`, strings.TrimSpace(code)).Scan(&resp.Code, &resp.Name)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (db *Database) AddAccount(acc *core.Account) error {
	log.Debug("Adding Account to DB")
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
	u, _ := db.FindAccount(strings.TrimSpace(acc.Code))
	if u != nil {
		return nil
	}
	return db.AddAccount(acc)

}

func (db *Database) DeleteAccount(account string) error {

	sqlStatement := `
	DELETE FROM accounts
	WHERE 
		name = ?
	;`
	_, err := db.DB.Exec(sqlStatement, account)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) FindUser(pubKey string) (*core.User, error) {
	var resp core.User
	log.Debug("Searching User in DB")
	err := db.DB.QueryRow(`SELECT * FROM users WHERE username = ? LIMIT 1`, pubKey).Scan(&resp.Id, &resp.Name)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (db *Database) AddUser(usr *core.User) error {
	log.Debug("Adding User to DB")
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
	log.Debug("Testing DB")
	createDB := "create table if not exists pages (title text, body blob, timestamp text)"
	log.Debug("Query: " + createDB)
	res, err := db.DB.Exec(createDB)
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

	tx, _ := db.DB.Begin()

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	stmt, _ := tx.Prepare("insert into pages (title, body, timestamp) values (?, ?, ?)")
	log.Debug("Query: Insert")
	res, err = stmt.Exec("Sean", "Body", timestamp)
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

	return err
}

func (db *Database) GetTB(queryDate time.Time) (*[]core.TBAccount, error) {

	queryDB := `
		SELECT split_accounts.account_id,
					 Sum(splits.amount),
					 splits.currency,
					 currencies.decimals
		FROM   splits
					 JOIN split_accounts
						 ON splits.split_id = split_accounts.split_id
					 JOIN currencies
						 ON splits.currency = currencies.name
		WHERE  splits.split_date <= ?
					 AND "void" NOT IN (SELECT t.tag_name
															FROM   tags AS t
																		 JOIN transaction_tag AS tt
																			 ON tt.tag_id = t.tag_id
															WHERE  tt.transaction_id = splits.transaction_id)
		GROUP  BY split_accounts.account_id, splits.currency
		;`

	log.Debug("Querying Database for Trial Balance")

	rows, err := db.DB.Query(queryDB, queryDate)
	if err != nil {
		log.Fatal("Trial Balance Query Failed with error: ", err)
	}
	defer rows.Close()

	accounts := []core.TBAccount{}

	for rows.Next() {
		var t core.TBAccount
		if err := rows.Scan(&t.Account, &t.Amount, &t.Currency, &t.Decimals); err != nil {
			log.Fatal(err)
		}
		accounts = append(accounts, t)
	}
	if rows.Err() != nil {
		log.Fatal(err)
	}

	tagsQuery := `
		SELECT tag_name
		FROM   tags
					 JOIN account_tag
						 ON account_tag.tag_id = tags.tag_id
					 JOIN accounts
						 ON accounts.account_id = account_tag.account_id
		WHERE  accounts.NAME = ?;
		`

	for index, element := range accounts {
		rows, err = db.DB.Query(tagsQuery, element.Account)
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			var tag string
			if err := rows.Scan(&tag); err != nil {
				log.Fatal(err)
			}
			accounts[index].Tags = append(accounts[index].Tags, tag)
		}

	}

	return &accounts, nil

}

func (db *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.Query(query, args...)

}

func (db *Database) GetListing(startDate, endDate time.Time) (*[]core.Transaction, error) {

	var txns []core.Transaction

	log.Debugf("Searching Transactions in DB between %s & %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// Find the transaction bodys
	rows, err := db.DB.Query(`
		SELECT
        t.transaction_id
        ,t.postdate
        ,t.brief
        ,u.user_id
        ,u.username
    FROM
        transactions AS t JOIN users AS u
            ON t.poster_user_id = u.user_id
			;`)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var t core.Transaction
		var poster core.User

		if err := rows.Scan(&t.Id, &t.Postdate, &t.Description, &poster.Id, &poster.Name); err != nil {
			log.Fatal(err)
		}

		// Find all splits relating to that transaction
		splits, err := db.Query(`
				SELECT s.split_id,
							 s.split_date,
							 s.description,
							 a.account_id,
							 a.NAME,
							 s.currency,
							 c.decimals,
							 s.amount
				FROM   splits AS s
							 JOIN split_accounts AS sa
								 ON s.split_id = sa.split_id
							 JOIN accounts AS a
								 ON sa.account_id = a. account_id
							 JOIN currencies AS c
								 ON s.currency = c.NAME
				WHERE  s.transaction_id = ?
        AND    s.split_date BETWEEN ? AND ?
				;`,
			t.Id,
			startDate.Format("2006-01-02"),
			endDate.Format("2006-01-02"))
		if err != nil {
			return nil, err
		}

		for splits.Next() {
			var split core.Split
			var account core.Account
			var cur core.Currency
			var amount int64
			// for each row, scan the result into our split object
			err = splits.Scan(&split.Id, &split.Date, &split.Description, &account.Code, &account.Name, &cur.Name, &cur.Decimals, &amount)
			if err != nil {
				return nil, err
			}
			split.Amount = big.NewInt(amount)
			split.Accounts = append(split.Accounts, &account)
			split.Currency = &cur
			t.Splits = append(t.Splits, &split)

		}
		if len(t.Splits) > 0 {
			txns = append(txns, t)
		}
	}
	if rows.Err() != nil {
		log.Fatal(err)
	}

	return &txns, nil
}
