package main

import (
	"flag"
	"math/big"
	"testing"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/darcys22/godbledger/godbledger/cmd"
	"github.com/darcys22/godbledger/godbledger/core"
	"github.com/darcys22/godbledger/godbledger/ledger"
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

	set := flag.NewFlagSet("test", 0)
	set.String("config", "", "doc")

	ctx := cli.NewContext(nil, set, nil)

	err, cfg := cmd.MakeConfig(ctx)
	if err != nil {
		t.Fatalf("New Config Failed: %v", err)
	}

	cfg.DatabaseType = "sqlite3"
	cfg.DatabaseLocation = ":memory:"

	ledger, err := ledger.New(ctx, cfg)
	if err != nil {
		t.Fatalf("New ledger Failed: %v", err)
	}
	ledger.Start()

	//response, err := ledger.Insert(txn)
	_, err = ledger.Insert(txn)
	if err != nil {
		t.Fatalf("Inserting to ledger Failed: %v", err)
	}

	//_, valid := ledger.Transactions[0].Balance()
	//if !valid {
	//t.Fatalf("Invalid Transaction")
	//}

	ledger.Stop()
}
