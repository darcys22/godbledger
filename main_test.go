package main

import (
	"testing"
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
	aud, _ := NewCurrency("AUD",2)

	spl1 := NewSplit(time.Now(), "Cash Income", []Account{cash}, aud, 10)

	ledger, err := NewLedgerDB()
	if err != nil {
		t.Fatalf("New ledger Failed: %v", err)
	}

	err = ledger.AppendTransaction(txn)
	if err != nil {
		t.Fatalf("Appending to ledger Failed: %v", err)
	}

	if !ledger.Transactions[0].Valid() {
		t.Fatalf("Invalid Transaction")
	}

}
