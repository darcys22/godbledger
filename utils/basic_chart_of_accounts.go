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
	{AccountName: "Cash",
		Tags: []string{"main", "Current Asset", "Asset", "Balance Sheet"}},
	{AccountName: "Accounts Receivable",
		Tags: []string{"main", "Current Asset", "Asset", "Balance Sheet"}},
	{AccountName: "Accounts Payable",
		Tags: []string{"main", "Current Liability", "Liability", "Balance Sheet"}},
	{AccountName: "Retained Earnings",
		Tags: []string{"main", "Equity", "Balance Sheet"}},
	{AccountName: "Sales",
		Tags: []string{"main", "Revenue", "Profit and Loss"}},
	{AccountName: "General Expenses",
		Tags: []string{"main", "Expense", "Profit and Loss"}},
	{AccountName: "Rent",
		Tags: []string{"main", "Expense", "Profit and Loss"}},
	{AccountName: "Interest",
		Tags: []string{"main", "Expense", "Profit and Loss"}},
	{AccountName: "Computer Expenses",
		Tags: []string{"main", "Expense", "Profit and Loss"}},
	{AccountName: "Salary and Wages",
		Tags: []string{"main", "Expense", "Profit and Loss"}},
	{AccountName: "Minor Equipment",
		Tags: []string{"main", "Expense", "Profit and Loss"}},
	{AccountName: "Repairs and Maintenance",
		Tags: []string{"main", "Expense", "Profit and Loss"}},
	{AccountName: "Bank Account",
		Tags: []string{"External"}},
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
