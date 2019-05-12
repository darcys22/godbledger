package ledger

import (
	"godbledger/core"
)

type LedgerDB struct {
	Name         []byte
	Users        []core.User
	Currencies   []core.Currency
	Chart        []core.Account
	Transactions []core.Transaction
	Splits       []core.Split
}

func NewLedgerDB() (*LedgerDB, error) {
	//db, err := leveldb.OpenFile(file, &opt.Options{

	ldb := &LedgerDB{}
	return ldb, nil
}

func (db *LedgerDB) AppendTransaction(txn *core.Transaction) error {
	db.Transactions = append(db.Transactions, *txn)
	return nil
}

func (db *LedgerDB) AppendUser(usr core.User) error {
	db.Users = append(db.Users, usr)
	return nil
}
