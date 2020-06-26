package main

import (
	"context"
	"log"
	"strconv"
	"time"

	pb "github.com/darcys22/godbledger/proto"
	"google.golang.org/grpc"

	"github.com/urfave/cli"
)

var commandAddCurrency = &cli.Command{
	Name:      "currency",
	Usage:     "ledger_cli currency <currency name> <decimals>",
	ArgsUsage: "[]",
	Description: `
	Adds the tag specified in the second argument to the account specified in the first argument

	Example

	ledger_cli currency BTC 9
`,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "delete",
			Aliases: []string{"d"},
			Usage:   "deletes currency rather than creates",
		},
	},
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

			if c.Bool("delete") {
				req := &pb.DeleteCurrencyRequest{
					Currency:  c.Args().Get(0),
					Signature: "blah",
				}

				r, err := client.DeleteCurrency(ctx, req)
				if err != nil {
					log.Fatalf("could not greet: %v", err)
				}

				log.Printf("Delete Currency Response: %s", r.GetMessage())
			} else {

				if c.NArg() > 1 {

					decimals, err := strconv.ParseInt(c.Args().Get(1), 0, 64)
					if err != nil {
						log.Fatalf("could not parse the decimals provided: %v", err)
					}
					req := &pb.CurrencyRequest{
						Currency:  c.Args().Get(0),
						Decimals:  decimals,
						Signature: "blah",
					}

					r, err := client.AddCurrency(ctx, req)
					if err != nil {
						log.Fatalf("could not create currency: %v", err)
					}

					log.Printf("Create Currency Response: %s", r.GetMessage())

				} else {
					log.Printf("This command two arguments.")
				}
			}

			if err != nil {
				log.Fatalf("Failed with Error: %v", err)
			}

		} else {
			log.Printf("This command requires an argument.")
		}

		return nil
	},
}
