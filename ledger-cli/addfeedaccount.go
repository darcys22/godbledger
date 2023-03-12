package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/darcys22/godbledger/proto/transaction"

	"github.com/darcys22/godbledger/godbledger/cmd"

	"google.golang.org/grpc"

	"github.com/urfave/cli/v2"
)

var commandAddFeedAccount = &cli.Command{
	Name:      "addfeedaccount",
	Usage:     "ledger-cli addfeedaccount <account name> <currency>",
	ArgsUsage: "[]",
	Description: `
	Adds the account specified to the bank feeds accounts list, will not show up in trial balance but used for reconciliations

	Example

	ledger-cli addfeedaccount nab-account AUD
`,
	Flags: []cli.Flag{
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

      if ctx.NArg() >= 2 {
        req := &transaction.TransactionFeedAccountRequest{
          Name: ctx.Args().Get(0),
          Currency: ctx.Args().Get(1),
        }

        r, err := client.AddTransactionFeedAccount(ctxtimeout, req)
        if err != nil {
          return fmt.Errorf("Could not call Add Transaction Feed Account Method (%v)", err)
        }

        log.Infof("Add Transaction Feed Account Response: %s", r.GetMessage())

      } else {
        return errors.New("This command requires two arguments")
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
