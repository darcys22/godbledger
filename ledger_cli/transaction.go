package main

import (
	"context"
	"log"
	"math/big"
	"time"

	pb "github.com/darcys22/godbledger/proto"
	"google.golang.org/grpc"

	"github.com/urfave/cli/v2"
)

var commandSingleTestTransaction = &cli.Command{
	Name:      "single",
	Usage:     "submits a single transaction",
	ArgsUsage: "[]",
	Description: `
`,
	Flags: []cli.Flag{},
	Action: func(c *cli.Context) error {

		date, _ := time.Parse("2006-01-02", "2011-03-15")
		desc := "Whole Food Market"

		transactionLines := make([]Account, 2)

		line1Account := "Expenses:Groceries"
		line1Desc := "Groceries"
		line1Amount := big.NewRat(7500, 1)

		transactionLines[0] = Account{
			Name:        line1Account,
			Description: line1Desc,
			Balance:     line1Amount,
			Currency:    "USD",
		}

		line2Account := "Assets:Checking"
		line2Desc := "Groceries"
		line2Amount := big.NewRat(-7500, 1)

		transactionLines[1] = Account{
			Name:        line2Account,
			Description: line2Desc,
			Balance:     line2Amount,
			Currency:    "USD",
		}

		req := &Transaction{
			Date:           date,
			Payee:          desc,
			AccountChanges: transactionLines,
			Signature:      "stuff",
		}

		err := Send(req)
		if err != nil {
			log.Fatalf("could not send: %v", err)
		}

		return nil
	},
}

func Send(t *Transaction) error {

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewTransactorClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	transactionLines := make([]*pb.LineItem, 2)

	for i, accChange := range t.AccountChanges {
		transactionLines[i] = &pb.LineItem{
			Accountname: accChange.Name,
			Description: accChange.Description,
			Amount:      accChange.Balance.Num().Int64(),
			Currency:    accChange.Currency,
		}
	}

	req := &pb.TransactionRequest{
		Date:        t.Date.Format("2006-01-02"),
		Description: t.Payee,
		Lines:       transactionLines,
		Signature:   t.Signature,
	}
	r, err := client.AddTransaction(ctx, req)
	if err != nil {
		log.Fatalf("Could not send transaction: %v", err)
	}
	log.Printf("Response: %s", r.GetMessage())
	return nil
}
