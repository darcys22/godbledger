package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/darcys22/godbledger/godbledger/cmd"

	pb "github.com/darcys22/godbledger/proto"
	"google.golang.org/grpc"

	//"github.com/urfave/cli"
	"github.com/urfave/cli/v2"
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
	Action: func(ctx *cli.Context) error {
		err, cfg := cmd.MakeConfig(ctx)
		if err != nil {
			log.Fatal(err)
		}

		if ctx.NArg() > 0 {

			// Set up a connection to the server.
			address := fmt.Sprintf("%s:%s", cfg.Host, cfg.RPCPort)
			log.WithField("address", address).Info("GRPC Dialing on port")
			conn, err := grpc.Dial(address, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("did not connect: %v", err)
			}
			defer conn.Close()
			client := pb.NewTransactorClient(conn)

			ctxtimeout, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			if ctx.Bool("delete") {
				req := &pb.DeleteCurrencyRequest{
					Currency:  ctx.Args().Get(0),
					Signature: "blah",
				}

				r, err := client.DeleteCurrency(ctxtimeout, req)
				if err != nil {
					log.Fatalf("could not greet: %v", err)
				}

				log.Printf("Delete Currency Response: %s", r.GetMessage())
			} else {

				if ctx.NArg() > 1 {

					decimals, err := strconv.ParseInt(ctx.Args().Get(1), 0, 64)
					if err != nil {
						log.Fatalf("could not parse the decimals provided: %v", err)
					}
					req := &pb.CurrencyRequest{
						Currency:  ctx.Args().Get(0),
						Decimals:  decimals,
						Signature: "blah",
					}

					r, err := client.AddCurrency(ctxtimeout, req)
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
