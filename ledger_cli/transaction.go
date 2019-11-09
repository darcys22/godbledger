package main

import (
	"context"
	"log"
	"time"

	pb "github.com/darcys22/godbledger/proto"
	"google.golang.org/grpc"

	"github.com/urfave/cli"
)

var commandSingleTestTransaction = cli.Command{
	Name:      "single",
	Usage:     "submits a single transaction",
	ArgsUsage: "[]",
	Description: `
`,
	Flags: []cli.Flag{},
	Action: func(c *cli.Context) error {

		// Set up a connection to the server.
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewTransactorClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		date := "2011-03-15"
		desc := "Whole Food Market"

		transactionLines := make([]*pb.LineItem, 2)

		line1Account := "Expenses:Groceries"
		line1Desc := "Groceries"
		line1Amount := int64(7500)

		transactionLines[0] = &pb.LineItem{
			Accountname: line1Account,
			Description: line1Desc,
			Amount:      line1Amount,
		}

		line2Account := "Assets:Checking"
		line2Desc := "Groceries"
		line2Amount := int64(-7500)

		transactionLines[1] = &pb.LineItem{
			Accountname: line2Account,
			Description: line2Desc,
			Amount:      line2Amount,
		}

		req := &pb.TransactionRequest{
			Date:        date,
			Description: desc,
			Lines:       transactionLines,
		}
		r, err := client.AddTransaction(ctx, req)
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Version: %s", r.GetMessage())
		return nil
	},
}

func send(t *Transaction) error {

	// Set up a connection to the server.
	//conn, err := grpc.Dial(address, grpc.WithInsecure())
	//if err != nil {
	//log.Fatalf("did not connect: %v", err)
	//}
	//defer conn.Close()
	//client := pb.NewTransactorClient(conn)

	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()

	//transactionLines := make([]*pb.LineItem, 2)

	//for _, accChange := range t.AccountChanges {
	//transactionLine := make(*pb.LineItem)

	//}

	//req := &pb.TransactionRequest{
	//Date:        t.Date,
	//Description: desc,
	//Lines:       transactionLines,
	//}
	//r, err := client.AddTransaction(ctx, req)
	//if err != nil {
	//log.Fatalf("could not greet: %v", err)
	//}
	//log.Printf("Version: %s", r.GetMessage())
	return nil
}
