// This script will build a basic chart of accounts in a running godbledger server

package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/darcys22/godbledger/proto/transaction"

	"github.com/darcys22/godbledger/godbledger/cmd"

	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

type Account struct {
	AccountName string
	Tags        []string
}

var accounts = []Account{
	Account{AccountName: "Cash",
		Tags: []string{"Current Asset", "Asset", "Balance Sheet"}},
	Account{AccountName: "Accounts Receivable",
		Tags: []string{"Current Asset", "Asset", "Balance Sheet"}},
	Account{AccountName: "Accounts Payable",
		Tags: []string{"Current Liability", "Liability", "Balance Sheet"}},
	Account{AccountName: "Retained Earnings",
		Tags: []string{"Equity", "Balance Sheet"}},
	Account{AccountName: "Sales",
		Tags: []string{"Revenue", "Profit and Loss"}},
	Account{AccountName: "General Expenses",
		Tags: []string{"Expense", "Profit and Loss"}},
	Account{AccountName: "Rent",
		Tags: []string{"Expense", "Profit and Loss"}},
	Account{AccountName: "Interest",
		Tags: []string{"Expense", "Profit and Loss"}},
	Account{AccountName: "Computer Expenses",
		Tags: []string{"Expense", "Profit and Loss"}},
	Account{AccountName: "Salary and Wages",
		Tags: []string{"Expense", "Profit and Loss"}},
	Account{AccountName: "Minor Equipment",
		Tags: []string{"Expense", "Profit and Loss"}},
	Account{AccountName: "Repairs and Maintenance",
		Tags: []string{"Expense", "Profit and Loss"}},
}

func main() {

	// Create a config from the defaults which would usually be created by the CLI library
	set := flag.NewFlagSet("accounts", 0)
	set.String("config", "", "doc")
	ctx := cli.NewContext(nil, set, nil)
	err, config := cmd.MakeConfig(ctx)
	if err != nil {
		log.Fatalf("New Config Failed: %v", err)
	}

	conns := make([]*grpc.ClientConn, 1)
	for i := 0; i < len(conns); i++ {
		log.Printf("Starting GoDBLedger %d", i)
		conn, err := grpc.Dial(fmt.Sprintf("%s:%s", config.Host, config.RPCPort), grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Failed to dial: %v", err)
		}
		conns[i] = conn
		defer func() {
			if err := conn.Close(); err != nil {
				log.Fatal(err)
			}
		}()
	}

	client := transaction.NewTransactorClient(conns[0])

	for i := 0; i < len(accounts); i++ {
		req := &transaction.AccountTagRequest{
			Account: accounts[i].AccountName,
			Tag:     accounts[i].Tags,
		}
		_, err = client.AddAccount(context.Background(), req)
		if err != nil {
			log.Fatal(err)
		}
	}

}
