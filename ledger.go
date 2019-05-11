package main

import ()

type LedgerDB struct {
	Name         []byte
	Users        []User
	Currencies   []Currency
	Chart        []Account
	Transactions []Transaction
	Splits       []Split
}

func NewLedgerDB() (*LedgerDB, error) {
	//db, err := leveldb.OpenFile(file, &opt.Options{

	ldb := &LedgerDB{}
	return ldb, nil
}

func (db *LedgerDB) AppendTransaction(txn *Transaction) error {
	db.Transactions = append(db.Transactions, *txn)
	return nil
}

func (db *LedgerDB) AppendUser(usr User) error {
	db.Users = append(db.Users, usr)
	return nil
}
