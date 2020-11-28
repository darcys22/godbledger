package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/darcys22/godbledger/proto/transaction"

	"github.com/darcys22/godbledger/godbledger/cmd"

	"google.golang.org/grpc"

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
			return fmt.Errorf("Could not make config (%v)", err)
		}

		if ctx.NArg() > 0 {

			// Set up a connection to the server.
			address := fmt.Sprintf("%s:%s", cfg.Host, cfg.RPCPort)
			log.WithField("address", address).Info("GRPC Dialing on port")
			conn, err := grpc.Dial(address, grpc.WithInsecure())
			if err != nil {
				return fmt.Errorf("Could not connect to GRPC (%v)", err)
			}
			defer conn.Close()
			client := transaction.NewTransactorClient(conn)

			ctxtimeout, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			if ctx.Bool("delete") {
				req := &transaction.DeleteCurrencyRequest{
					Currency: ctx.Args().Get(0),
				}

				r, err := client.DeleteCurrency(ctxtimeout, req)
				if err != nil {
					return fmt.Errorf("Could not call Delete Currency Method (%v)", err)
				}

				log.Infof("Delete Currency Response: %s", r.GetMessage())
			} else {

				if ctx.NArg() > 1 {

					decimals, err := strconv.ParseInt(ctx.Args().Get(1), 0, 64)
					if err != nil {
						return fmt.Errorf("Could not parse the decimals provided (%v)", err)
					}
					req := &transaction.CurrencyRequest{
						Currency: ctx.Args().Get(0),
						Decimals: decimals,
					}

					r, err := client.AddCurrency(ctxtimeout, req)
					if err != nil {
						return fmt.Errorf("Could not call Add Currency Method (%v)", err)
					}

					log.Infof("Create Currency Response: %s", r.GetMessage())

				} else {
					return errors.New("This command requires two arguments")
				}
			}

			if err != nil {
				return fmt.Errorf("Failed with Error (%v)", err)
			}

		} else {
			return errors.New("This command requires an argument")
		}

		return nil
	},
}
