package main

import (
	"fmt"
	"time"

	"math/big"
)

type User struct {
	Name []byte
}

type Currency struct {
	Name     []byte
	Decimals int
}

type Account struct {
	Id   []byte
	Name []byte
}

type Transaction struct {
	Id          []byte
	Postdate    *time.Time
	Poster      User
	Splits      []*Split
	Description []byte
}

type Split struct {
	Transaction *Transaction

	Id          []byte
	Date        *time.Time
	Description []byte
	Accounts    []Account
	Currency    Currency
	Amount      big.Int
}
