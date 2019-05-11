package main

import (
	"math/big"
	"testing"
	"time"
)

func TestNewTransaction(t *testing.T) {

	user, err := NewUser("Tester")
	if err != nil {
		t.Fatalf("New User Failed: %v", err)
	}
	txn, err := NewTransaction(user)
	if err != nil {
		t.Fatalf("New Transaction Failed: %v", err)
	}

	cash, _ := NewAccount("1", "cash")
	income, _ := NewAccount("2", "income")
	aud, _ := NewCurrency("AUD", 2)

	spl1, err := NewSplit(time.Now(), []byte("Cash Income"), []*Account{cash}, aud, big.NewInt(10))
	if err != nil {
		t.Fatalf("Creating First Split Failed: %v", err)
	}

	err = txn.AppendSplit(spl1)
	if err != nil {
		t.Fatalf("Appending First Split Failed: %v", err)
	}

	spl2, err := NewSplit(time.Now(), []byte("Cash Income"), []*Account{income}, aud, big.NewInt(-10))
	if err != nil {
		t.Fatalf("Creating Second Split Failed: %v", err)
	}

	err = txn.AppendSplit(spl2)
	if err != nil {
		t.Fatalf("Appending Second Split Failed: %v", err)
	}

	ledger, err := NewLedgerDB()
	if err != nil {
		t.Fatalf("New ledger Failed: %v", err)
	}

	err = ledger.AppendTransaction(txn)
	if err != nil {
		t.Fatalf("Appending to ledger Failed: %v", err)
	}

	_, valid := ledger.Transactions[0].Balance()
	if !valid {
		t.Fatalf("Invalid Transaction")
	}

}
