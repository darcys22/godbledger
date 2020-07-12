package main

import (
	"context"
	"log"
	"time"

	pb "github.com/darcys22/godbledger/proto"
	"google.golang.org/grpc"

	//"github.com/urfave/cli"
	"github.com/urfave/cli/v2"
)

var commandDeleteTransaction = &cli.Command{
	Name:      "delete",
	Usage:     "ledger_cli delete <transaction_id>",
	ArgsUsage: "[]",
	Description: `
	Deletes a transaction from the database
`,
	Flags: []cli.Flag{},
	Action: func(c *cli.Context) error {

		if c.NArg() > 0 {
			// Set up a connection to the server.
			conn, err := grpc.Dial(address, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("did not connect: %v", err)
			}
			defer conn.Close()
			client := pb.NewTransactorClient(conn)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			signature := "test"
			req := &pb.DeleteRequest{
				Identifier: c.Args().Get(0),
				Signature:  signature,
			}
			r, err := client.DeleteTransaction(ctx, req)
			if err != nil {
				log.Fatalf("could not delete: %v", err)
			}
			log.Printf("Response: %s", r.GetMessage())
			return nil
		} else {
			log.Fatalf("This command requires an argument.")
		}

		return nil
	},
}

var commandVoidTransaction = &cli.Command{
	Name:      "void",
	Usage:     "ledger_cli void <transaction_id>",
	ArgsUsage: "[]",
	Description: `
	Reverses a transaction by creating an identical inverse and tags both transactions as void 
`,
	Flags: []cli.Flag{},
	Action: func(c *cli.Context) error {

		if c.NArg() > 0 {
			// Set up a connection to the server.
			conn, err := grpc.Dial(address, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("did not connect: %v", err)
			}
			defer conn.Close()
			client := pb.NewTransactorClient(conn)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			signature := "test"
			req := &pb.DeleteRequest{
				Identifier: c.Args().Get(0),
				Signature:  signature,
			}
			r, err := client.VoidTransaction(ctx, req)
			if err != nil {
				log.Fatalf("could not void: %v", err)
			}
			log.Printf("Response: %s", r.GetMessage())
			return nil
		} else {
			log.Fatalf("This command requires an argument.")
		}

		return nil
	},
}
