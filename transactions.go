package main

import (
	"github.com/rs/xid"
	"time"

	"math/big"
)

type User struct {
	Id   string
	Name string
}

func NewUser(name string) (*User, error) {
	guid := xid.New()

	usr := &User{guid.String(), name}
	return usr, nil

}

type Currency struct {
	Name     string
	Decimals int
}

func NewCurrency(name string, decimals int) (*Currency, error) {

	cur := &Currency{name, decimals}
	return cur, nil

}

type Account struct {
	Code string
	Name string
}

func NewAccount(code, name string) (*Account, error) {
	acc := &Account{code, name}
	return acc, nil
}

type Transaction struct {
	Id          string
	Postdate    time.Time
	Poster      *User
	Description []byte
	Splits      []*Split
}

func NewTransaction(usr *User) (*Transaction, error) {
	guid := xid.New()
	txn := &Transaction{guid.String(), time.Now(), usr, []byte{}, []*Split{}}
	return txn, nil
}

func (txn *Transaction) Valid() bool {
	return false
}

type Split struct {
	Id          string
	Date        *time.Time
	Description []byte
	Accounts    []Account
	Currency    Currency
	Amount      big.Int
}

func NewSplit(date *time.Time, desc []byte, accs []Account, cur Currency, amt big.Int) (*Split, error) {
	guid := xid.New()
	spl := &Split{guid.String(), date, desc, accs, cur, amt}
	return spl, nil
}
