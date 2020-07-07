package main

import (
	"github.com/darcys22/godbledger/godbledger/core"
	//"github.com/darcys22/godbledger/godbledger/ledger"
	"math/big"
	"testing"
	"time"
)

func TestNewTransaction(t *testing.T) {

	user, err := core.NewUser("Tester")
	if err != nil {
		t.Fatalf("New User Failed: %v", err)
	}
	txn, err := core.NewTransaction(user)
	if err != nil {
		t.Fatalf("New Transaction Failed: %v", err)
	}

	cash, _ := core.NewAccount("1", "cash")
	income, _ := core.NewAccount("2", "income")
	aud, _ := core.NewCurrency("AUD", 2)

	spl1, err := core.NewSplit(time.Now(), []byte("Cash Income"), []*core.Account{cash}, aud, big.NewInt(10))
	if err != nil {
		t.Fatalf("Creating First Split Failed: %v", err)
	}

	err = txn.AppendSplit(spl1)
	if err != nil {
		t.Fatalf("Appending First Split Failed: %v", err)
	}

	spl2, err := core.NewSplit(time.Now(), []byte("Cash Income"), []*core.Account{income}, aud, big.NewInt(-10))
	if err != nil {
		t.Fatalf("Creating Second Split Failed: %v", err)
	}

	err = txn.AppendSplit(spl2)
	if err != nil {
		t.Fatalf("Appending Second Split Failed: %v", err)
	}

	//ledger, err := ledger.NewLedgerDB()
	//if err != nil {
	//t.Fatalf("New ledger Failed: %v", err)
	//}

	//err = ledger.AppendTransaction(txn)
	//if err != nil {
	//t.Fatalf("Appending to ledger Failed: %v", err)
	//}

	//_, valid := ledger.Transactions[0].Balance()
	//if !valid {
	//t.Fatalf("Invalid Transaction")
	//}

}
